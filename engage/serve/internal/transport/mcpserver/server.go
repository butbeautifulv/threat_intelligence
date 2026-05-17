package mcpserver

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/butbeautifulv/veil/engage/serve/internal/ports"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/bugbounty"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tooldispatch"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/version"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

type Server struct {
	dispatch    *tooldispatch.Dispatcher
	runner      *toolsuc.Runner
	intel       ports.IntelProvider
	cve         ports.CVEProvider
	ctf         ports.CTFProvider
	bugbounty   *bugbounty.Service
	browser     *browser.Service
	processes   *process.Manager
	workflows   *workflow.Service
	auth        *auth.Stack
	logger      *slog.Logger
	catalogPath string
	files       *files.Manager
}

func NewServer(runner *toolsuc.Runner, stack *auth.Stack, logger *slog.Logger) *Server {
	return &Server{
		dispatch: tooldispatch.NewDispatcher(runner, nil, nil, nil, nil, nil, nil, nil, "", nil),
		runner:   runner,
		auth:     stack,
		logger:   logger,
	}
}

// NewServerWithIntel wires in-process intelligence and workflow handlers for MCP tools/call.
func NewServerWithIntel(runner *toolsuc.Runner, intel ports.IntelProvider, wf *workflow.Service, stack *auth.Stack, logger *slog.Logger, catalogPath string, fileMgr *files.Manager) *Server {
	return &Server{
		dispatch:    tooldispatch.NewDispatcher(runner, intel, nil, nil, nil, nil, nil, wf, catalogPath, fileMgr),
		runner: runner, intel: intel, workflows: wf, auth: stack, logger: logger, catalogPath: catalogPath, files: fileMgr,
	}
}

// NewServerWithCTF wires intelligence, CTF, and workflow handlers.
func NewServerWithCTF(runner *toolsuc.Runner, intel ports.IntelProvider, ctfSvc ports.CTFProvider, wf *workflow.Service, stack *auth.Stack, logger *slog.Logger, catalogPath string, fileMgr *files.Manager) *Server {
	return &Server{
		dispatch:    tooldispatch.NewDispatcher(runner, intel, nil, ctfSvc, nil, nil, nil, wf, catalogPath, fileMgr),
		runner: runner, intel: intel, ctf: ctfSvc, workflows: wf, auth: stack, logger: logger, catalogPath: catalogPath, files: fileMgr,
	}
}

// NewServerFull wires intelligence, CVE, CTF, bug bounty, browser, processes, and workflow handlers.
func NewServerFull(runner *toolsuc.Runner, intel ports.IntelProvider, cveSvc ports.CVEProvider, ctfSvc ports.CTFProvider, bb *bugbounty.Service, browserSvc *browser.Service, proc *process.Manager, wf *workflow.Service, stack *auth.Stack, logger *slog.Logger, catalogPath string, fileMgr *files.Manager) *Server {
	return &Server{
		dispatch:    tooldispatch.NewDispatcher(runner, intel, cveSvc, ctfSvc, bb, browserSvc, proc, wf, catalogPath, fileMgr),
		runner: runner, intel: intel, cve: cveSvc, ctf: ctfSvc, bugbounty: bb, browser: browserSvc, processes: proc, workflows: wf, auth: stack, logger: logger, catalogPath: catalogPath, files: fileMgr,
	}
}

// NewServerWithDispatch uses a shared dispatcher (e.g. from components.InitAPI).
func NewServerWithDispatch(dispatch *tooldispatch.Dispatcher, runner *toolsuc.Runner, stack *auth.Stack, logger *slog.Logger) *Server {
	return &Server{dispatch: dispatch, runner: runner, auth: stack, logger: logger}
}

func (s *Server) Run(ctx context.Context, inReader any, outWriter any) error {
	return mcp.RunStdio(ctx, s, inReader, outWriter)
}

func (s *Server) ProcessMessage(ctx context.Context, msg rpcMessage, httpTransport bool) (resp *rpcMessage, isNotification bool, err error) {
	result, rerr := s.handle(ctx, msg.Method, msg.Params, httpTransport)
	return mcp.BuildResponse(msg, result, rerr)
}

func (s *Server) handle(ctx context.Context, method string, params json.RawMessage, httpTransport bool) (any, error) {
	switch method {
	case "initialize":
		return mcp.InitializeResult(version.ServerName, version.MCP(), httpTransport, params), nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		return listToolsPayload(s.runner.Registry.ListAll()), nil

	case "tools/call":
		p, err := mcp.ParseToolCallParams(params)
		if err != nil {
			return nil, err
		}
		ctx, err = mcp.AuthorizeToolCall(ctx, s.auth, auth.PermEngageToolRun, auth.AuthorizeEngageMCP)
		if err != nil {
			return nil, err
		}
		return s.callTool(ctx, p.Name, p.Arguments)

	case "notifications/initialized":
		return nil, nil
	}

	return nil, rpcErrf(codeMethodNotFound, "unknown method: %s", method)
}

// CatalogCount exposes registry size for healthchecks.
func CatalogCount(reg *tools.Registry) int {
	if reg == nil {
		return 0
	}
	return reg.Count()
}
