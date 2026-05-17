package tooldispatch

import (
	"context"
	"errors"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func TestIsIntelBridgeTool(t *testing.T) {
	if !IsIntelBridgeTool("comprehensive_api_audit", tool.Spec{Category: toolid.CategoryWeb}) {
		t.Fatal("comprehensive_api_audit should bridge")
	}
	if !IsIntelBridgeTool("analyze_target_intelligence", tool.Spec{Category: toolid.CategoryIntel}) {
		t.Fatal("intelligence category should bridge")
	}
	if !IsIntelBridgeTool("monitor_cve_feeds", tool.Spec{Category: toolid.CategoryWeb}) {
		t.Fatal("monitor_cve_feeds should bridge")
	}
}

type mockNVD struct{}

func (mockNVD) FetchCVE(_ context.Context, cveID string) (*cve.CVEEntry, error) {
	return &cve.CVEEntry{
		CVEID:       cveID,
		Description: "SQL injection in login form",
		Severity:    "HIGH",
		CVSSScore:   8.1,
	}, nil
}

func (mockNVD) FetchRecent(_ context.Context, _ int, _ string) ([]cve.CVEEntry, error) {
	return []cve.CVEEntry{
		{CVEID: "CVE-2020-0001", Description: "xss reflected", Severity: "HIGH", CVSSScore: 7.5},
	}, nil
}

func TestDispatch_monitorCVE_withoutEnabled(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "monitor_cve_feeds", Category: toolid.CategoryIntel, Enabled: false},
	})
	runner := &toolsuc.Runner{Registry: reg}
	cveSvc := cve.NewService(nil, mockNVD{})
	d := NewDispatcher(runner, &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}, cveSvc, nil, nil, nil, nil, nil, "", nil)

	_, err := d.Dispatch(context.Background(), "", "monitor_cve_feeds", map[string]any{"hours": 24})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDispatch_unknownTool(t *testing.T) {
	reg := tools.NewRegistry(nil)
	d := NewDispatcher(&toolsuc.Runner{Registry: reg}, nil, nil, nil, nil, nil, nil, nil, "", nil)
	_, err := d.Dispatch(context.Background(), "", "no_such_tool", nil)
	var de *DispatchError
	if !errors.As(err, &de) || !de.NotFound {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestDispatch_getProcessDashboard(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "get_process_dashboard", Category: toolid.CategoryWeb, Enabled: false},
	})
	runner := &toolsuc.Runner{Registry: reg}
	proc := process.NewManager()
	d := NewDispatcher(runner, nil, nil, nil, nil, nil, proc, nil, "", nil)
	out, err := d.Dispatch(context.Background(), "", "get_process_dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("expected dashboard")
	}
}

func TestIsBridgeWorkflowBinary_pythonRunnerBinaries(t *testing.T) {
	for _, bin := range []string{"engage-python-install", "engage-python-exec"} {
		if IsBridgeWorkflowBinary(bin) {
			t.Fatalf("binary %q should run via subprocess runner, not workflow bridge", bin)
		}
	}
}

func TestDispatch_pythonTools_useRunnerNotBridge(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{
			Name:       "install_python_package",
			Binary:     "engage-python-install",
			Category:   toolid.CategoryWeb,
			Enabled:    true,
			TimeoutSec: 300,
			Parameters: []tool.Param{
				{Name: "target", Required: true},
				{Name: "package", Required: true},
				{Name: "env_name", Default: "default"},
			},
			ArgsTemplate: []string{"--env", "{env_name}", "--package", "{package}", "--target", "{target}"},
		},
		{
			Name:       "execute_python_script",
			Binary:     "engage-python-exec",
			Category:   toolid.CategoryWeb,
			Enabled:    true,
			TimeoutSec: 300,
			Parameters: []tool.Param{
				{Name: "target", Required: true},
				{Name: "script", Required: true},
				{Name: "env_name", Default: "default"},
				{Name: "filename", Default: ""},
			},
			ArgsTemplate: []string{
				"--env", "{env_name}", "--script", "{script}",
				"--filename", "{filename}", "--target", "{target}",
			},
		},
	})
	run := &toolsuc.Runner{Registry: reg, Exec: &runner.Executor{}}
	d := NewDispatcher(run, nil, nil, nil, nil, nil, nil, nil, "", nil)

	for _, tc := range []struct {
		tool string
		args map[string]any
	}{
		{
			tool: "install_python_package",
			args: map[string]any{"target": "local", "package": "six"},
		},
		{
			tool: "execute_python_script",
			args: map[string]any{"target": "local", "script": "print('engage-p10b')"},
		},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			out, err := d.Dispatch(context.Background(), "", tc.tool, tc.args)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := out.(map[string]any); ok {
				t.Fatalf("%s: expected runner ToolRunResponse, got workflow bridge map", tc.tool)
			}
			res, ok := out.(contract.ToolRunResponse)
			if !ok {
				t.Fatalf("%s: unexpected result type %T", tc.tool, out)
			}
			if res.Tool != tc.tool {
				t.Fatalf("tool = %q", res.Tool)
			}
			// Binary may be absent on dev PATH; runner path must still be taken.
			if res.Success {
				return
			}
			if res.Error == "" {
				t.Fatal("expected error when binary missing or command failed")
			}
		})
	}
}

func TestDispatch_analyzeTarget(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "analyze_target_intelligence", Category: toolid.CategoryIntel, Enabled: false},
	})
	runner := &toolsuc.Runner{Registry: reg}
	intel := &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}
	d := NewDispatcher(runner, intel, nil, nil, nil, nil, nil, nil, "", nil)

	out, err := d.Dispatch(context.Background(), "", "analyze_target_intelligence", map[string]any{
		"target": "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("expected analysis")
	}
}
