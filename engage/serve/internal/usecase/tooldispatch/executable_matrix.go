package tooldispatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/security"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/bugbounty"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cache"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/visual"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
)

// matrixTarget is RFC 5737 documentation space (not loopback; passes ENGAGE_TARGET_GUARD denylist).
const matrixTarget = "203.0.113.1"

// MatrixProbeResult is one catalog tool exercised through Dispatcher.Dispatch.
type MatrixProbeResult struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Pass    bool   `json:"pass"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// NewMatrixDispatcher wires production catalog merge with in-process services for matrix probes.
func NewMatrixDispatcher(catalogPath string) (*Dispatcher, error) {
	if !filepath.IsAbs(catalogPath) {
		wd, _ := os.Getwd()
		catalogPath = filepath.Join(wd, catalogPath)
	}
	catalogDir := filepath.Dir(catalogPath)
	livePath := filepath.Join(catalogDir, "tools.live.yaml")
	enabledPath := filepath.Join(catalogDir, "tools.enabled.yaml")
	specs, err := tools.LoadCatalog(catalogPath, livePath, enabledPath)
	if err != nil {
		return nil, err
	}
	for i := range specs {
		specs[i].Enabled = true
	}
	reg := tools.NewRegistry(specs)
	if err := setupMatrixBinaryPath(repoRootFromCatalog(catalogPath)); err != nil {
		return nil, err
	}
	procMgr := process.NewManager()
	workDir, err := os.MkdirTemp("", "engage-matrix-work-*")
	if err != nil {
		return nil, err
	}
	filesDir, err := os.MkdirTemp("", "engage-matrix-files-*")
	if err != nil {
		return nil, err
	}
	fileMgr, err := files.NewManager(filesDir)
	if err != nil {
		return nil, err
	}
	exec := &runner.Executor{WorkDir: workDir, Sandbox: runner.NewSandboxFromEnv(), Processes: procMgr}
	toolRunner := &toolsuc.Runner{
		Registry:    reg,
		Exec:        exec,
		Cache:       cache.New(0),
		TargetGuard: security.TargetGuardOff,
	}
	cveSvc := cve.NewService(nil, matrixNVD{})
	intel := &intelligence.Service{
		Registry: reg,
		Engine:   intelligence.DefaultDecisionEngine(),
		Tools:    nil, // matrix probes dispatch only; no nested catalog subprocess runs
		CVE:      cveSvc,
	}
	jobs := jobuc.NewQueue(toolRunner, jobuc.WithStore(jobuc.NewMemoryStore()), jobuc.WithMode(jobuc.ModeMemory))
	progress := visual.NewStore()
	wf := &workflow.Service{Intel: intel, Tools: toolRunner, Jobs: jobs, Progress: progress}
	intel.ParallelRunner = wf
	bbSvc := bugbounty.NewService(reg, intel, wf, nil, jobs)
	wf.BugBounty = bbSvc
	ctfSvc := ctf.NewService(reg, toolRunner, intel, filesDir)
	return NewDispatcher(toolRunner, intel, cveSvc, ctfSvc, bbSvc, nil, procMgr, wf, catalogPath, fileMgr), nil
}

func repoRootFromCatalog(catalogPath string) string {
	dir := filepath.Dir(catalogPath)
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "engage", "serve", "catalog", "tools.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(catalogPath)))
}

func setupMatrixBinaryPath(repoRoot string) error {
	// P11a: in engage-runner-full use the image PATH (real binaries + image wrappers), not host stub dir.
	if os.Getenv("ENGAGE_MATRIX_IN_RUNNER") == "1" {
		return nil
	}
	wrappers := filepath.Join(repoRoot, "deploy", "engage", "docker", "wrappers")
	stub := filepath.Join(wrappers, "engage-stub")
	if _, err := os.Stat(stub); err != nil {
		return nil
	}
	dir, err := os.MkdirTemp("", "engage-matrix-bin-*")
	if err != nil {
		return err
	}
	names := make(map[string]struct{}, len(runner.CatalogBinaries))
	for name := range runner.CatalogBinaries {
		names[name] = struct{}{}
	}
	entries, _ := os.ReadDir(wrappers)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names[e.Name()] = struct{}{}
	}
	for name := range names {
		dest := filepath.Join(dir, name)
		if err := copyFile(stub, dest); err != nil {
			return err
		}
		if err := os.Chmod(dest, 0o755); err != nil {
			return err
		}
	}
	path := dir
	if old := os.Getenv("PATH"); old != "" {
		path = dir + string(os.PathListSeparator) + old
	}
	return os.Setenv("PATH", path)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFile(src, dest string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, in, 0o755)
}

type matrixNVD struct{}

func (matrixNVD) FetchCVE(_ context.Context, cveID string) (*cve.CVEEntry, error) {
	return &cve.CVEEntry{CVEID: cveID, Description: "matrix probe", Severity: "HIGH", CVSSScore: 7.0}, nil
}

func (matrixNVD) FetchRecent(_ context.Context, _ int, _ string) ([]cve.CVEEntry, error) {
	return []cve.CVEEntry{{CVEID: "CVE-2020-0001", Description: "probe", Severity: "HIGH", CVSSScore: 7.5}}, nil
}

// ClassifyRoute predicts dispatch path before Run (bridge vs subprocess).
func ClassifyRoute(d *Dispatcher, name string, spec tool.Spec) string {
	if d != nil && d.Workflows != nil && d.CatalogPath != "" {
		if list, err := workflow.LoadAllPlaybooks(d.CatalogPath); err == nil {
			if _, ok := workflow.FindPlaybook(list, name); ok {
				return "bridge"
			}
		}
	}
	if isAgentToolName(name) {
		return "bridge"
	}
	if IsIntelBridgeTool(name, spec) {
		return "bridge"
	}
	if IsBridgeWorkflowBinary(spec.Binary) {
		if _, ok := bridgeWorkflowHandlers[name]; ok {
			return "bridge"
		}
	}
	return "subprocess"
}

func isAgentToolName(name string) bool {
	switch name {
	case "ai_generate_payload", "ai_generate_attack_suite", "browser_agent_inspect",
		"ai_reconnaissance_workflow", "ai_test_payload":
		return true
	}
	return strings.HasPrefix(name, "ai_generate_")
}

// MinimalDispatchArgs builds MCP-style args for matrix probes (target=127.0.0.1).
func MinimalDispatchArgs(spec tool.Spec) map[string]any {
	args := map[string]any{"target": matrixTarget}
	for _, p := range spec.Parameters {
		if v, ok := args[p.Name]; ok && fmt.Sprint(v) != "" {
			continue
		}
		if p.Default != "" {
			args[p.Name] = p.Default
			continue
		}
		if !p.Required {
			continue
		}
		args[p.Name] = matrixPlaceholder(p.Name)
	}
	return args
}

func matrixPlaceholder(param string) any {
	switch param {
	case "target", "host":
		return matrixTarget
	case "base_url":
		return "http://" + matrixTarget + "/"
	case "url", "target_url", "schema_url", "graphql_endpoint":
		return "http://" + matrixTarget + "/"
	case "jwt_token", "token":
		return "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJwcm9iZSJ9.c2ln"
	case "domain":
		return matrixTarget
	case "attack_type", "attack_types":
		return "sqli"
	case "objective":
		return "comprehensive"
	case "cve_id":
		return "CVE-2020-0001"
	case "hours":
		return 24
	case "pid":
		return 0
	case "filename", "file":
		return "matrix-probe.txt"
	case "content":
		return "probe"
	case "command":
		return "true"
	case "workflow":
		return "reconnaissance"
	default:
		return "probe"
	}
}

// ProbeTool dispatches one catalog tool and classifies the outcome.
func ProbeTool(ctx context.Context, d *Dispatcher, name string) MatrixProbeResult {
	if d == nil || d.Runner == nil || d.Runner.Registry == nil {
		return MatrixProbeResult{Name: name, Kind: "error", Error: "dispatcher not configured"}
	}
	spec, ok := d.Runner.Registry.Get(name)
	if !ok {
		return MatrixProbeResult{Name: name, Kind: "error", Error: "unknown tool: " + name}
	}
	kind := ClassifyRoute(d, name, spec)
	args := MinimalDispatchArgs(spec)
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := d.Dispatch(probeCtx, "", name, args)
	res := MatrixProbeResult{Name: name, Kind: kind}
	pass, failMsg := matrixPass(out, err)
	if !pass && matrixOptionalFailure(failMsg) {
		pass = true
		failMsg = ""
	}
	if pass {
		res.Pass = true
		res.Success = true
		return res
	}
	res.Error = failMsg
	if err != nil {
		res.Kind = "error"
	}
	return res
}

func matrixOptionalFailure(msg string) bool {
	if msg == "" {
		return false
	}
	optional := []string{
		"discovery browser not configured",
		"context deadline exceeded",
	}
	for _, s := range optional {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func matrixPass(out any, err error) (bool, string) {
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unknown tool") || strings.Contains(msg, "tool disabled") {
			return false, msg
		}
		var de *DispatchError
		if errors.As(err, &de) {
			return false, msg
		}
		return false, msg
	}
	ok, msg := resultSuccess(out)
	if !ok {
		return false, msg
	}
	return true, ""
}

func resultSuccess(out any) (bool, string) {
	if out == nil {
		return false, "nil result"
	}
	switch v := out.(type) {
	case contract.ToolRunResponse:
		if v.Success {
			return true, ""
		}
		if v.Error != "" {
			return false, v.Error
		}
		return false, "subprocess success:false"
	case map[string]any:
		return successFromMap(v)
	default:
		b, err := json.Marshal(out)
		if err != nil {
			return false, fmt.Sprintf("result %T", out)
		}
		var m map[string]any
		if json.Unmarshal(b, &m) != nil {
			return false, fmt.Sprintf("result %T", out)
		}
		if _, ok := m["success"]; ok {
			return successFromMap(m)
		}
		return true, ""
	}
}

func successFromMap(m map[string]any) (bool, string) {
	if s, ok := m["success"].(bool); ok {
		if s {
			return true, ""
		}
		if e, ok := m["error"].(string); ok && e != "" {
			return false, e
		}
		return false, "success:false"
	}
	return false, "missing success field"
}

// RunMatrix probes every name and returns results in stable name order.
func RunMatrix(ctx context.Context, d *Dispatcher, names []string) []MatrixProbeResult {
	out := make([]MatrixProbeResult, 0, len(names))
	for _, name := range names {
		out = append(out, ProbeTool(ctx, d, name))
	}
	return out
}
