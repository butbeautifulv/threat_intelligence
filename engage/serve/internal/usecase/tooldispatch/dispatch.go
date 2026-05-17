package tooldispatch

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/ports"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/bugbounty"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

// Dispatcher routes tool calls through playbook, agent, intel bridge, then subprocess runner.
type Dispatcher struct {
	Runner      *tools.Runner
	Intel       ports.IntelProvider
	CVE         ports.CVEProvider
	CTF         ports.CTFProvider
	BugBounty   *bugbounty.Service
	Browser     *browser.Service
	Processes   *process.Manager
	Workflows   *workflow.Service
	CatalogPath string
	Files       *files.Manager
}

// DispatchRequest runs a tool using the same order as MCP tools/call.
func (d *Dispatcher) DispatchRequest(ctx context.Context, subject, name string, req contract.ToolRunRequest) (any, error) {
	return d.Dispatch(ctx, subject, name, ArgsFromRequest(req))
}

// Dispatch runs a tool by name and returns a JSON-serializable result (not MCP content envelope).
func (d *Dispatcher) Dispatch(ctx context.Context, subject, name string, args map[string]any) (any, error) {
	if d == nil || d.Runner == nil {
		return nil, dispatchToolError("tool runner not configured")
	}
	if subject == "" {
		if sub, ok := auth.SubjectFromContext(ctx); ok {
			subject = sub.Sub
		}
	}
	target := argTarget(args)
	if out, ok, err := d.tryPlaybookByName(ctx, subject, name, target, argBool(args, "async")); ok {
		return out, err
	}
	if out, ok, err := d.tryAgentTool(ctx, name, args); ok {
		return out, err
	}
	spec, ok := d.Runner.Registry.Get(name)
	if !ok {
		return nil, dispatchNotFound("unknown tool: %s", name)
	}
	if IsIntelBridgeTool(name, spec) {
		return d.callIntelBridge(ctx, name, spec, args)
	}
	if IsBridgeWorkflowBinary(spec.Binary) {
		if out, ok, err := d.tryBridgeWorkflowTool(ctx, name, args); ok {
			return out, err
		}
	}
	res := d.Runner.Run(ctx, subject, name, RequestFromArgs(args))
	if !res.Success && res.Error != "" {
		return nil, dispatchToolError("%s", res.Error)
	}
	return res, nil
}

// IsIntelBridgeTool returns true when dispatch should use in-process intelligence handlers.
func IsIntelBridgeTool(name string, spec tool.Spec) bool {
	if name == "comprehensive_api_audit" || name == "target_timeline_intelligence" || name == "target_graph_context" {
		return true
	}
	if name == "monitor_cve_feeds" || name == "generate_exploit_from_cve" {
		return true
	}
	if spec.Category == toolid.CategoryCTF {
		return true
	}
	if strings.HasPrefix(name, "ctf_") {
		return true
	}
	return spec.Category == toolid.CategoryIntel
}

// NewDispatcher wires the unified tool dispatch path for HTTP and MCP.
func NewDispatcher(
	runner *tools.Runner,
	intel ports.IntelProvider,
	cve ports.CVEProvider,
	ctf ports.CTFProvider,
	bb *bugbounty.Service,
	browserSvc *browser.Service,
	proc *process.Manager,
	wf *workflow.Service,
	catalogPath string,
	fileMgr *files.Manager,
) *Dispatcher {
	return &Dispatcher{
		Runner:      runner,
		Intel:       intel,
		CVE:         cve,
		CTF:         ctf,
		BugBounty:   bb,
		Browser:     browserSvc,
		Processes:   proc,
		Workflows:   wf,
		CatalogPath: catalogPath,
		Files:       fileMgr,
	}
}
