package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/version"
	"github.com/butbeautifulv/veil/pkg/auth"
)

type Server struct {
	runner      *toolsuc.Runner
	intel       *intelligence.Service
	workflows   *workflow.Service
	auth        *auth.Stack
	logger      *slog.Logger
	catalogPath string
	files       *files.Manager
}

func NewServer(runner *toolsuc.Runner, stack *auth.Stack, logger *slog.Logger) *Server {
	return &Server{runner: runner, auth: stack, logger: logger}
}

// NewServerWithIntel wires in-process intelligence and workflow handlers for MCP tools/call.
func NewServerWithIntel(runner *toolsuc.Runner, intel *intelligence.Service, wf *workflow.Service, stack *auth.Stack, logger *slog.Logger, catalogPath string, fileMgr *files.Manager) *Server {
	return &Server{runner: runner, intel: intel, workflows: wf, auth: stack, logger: logger, catalogPath: catalogPath, files: fileMgr}
}

func (s *Server) Run(ctx context.Context, inReader any, outWriter any) error {
	in, ok := inReader.(interface{ Read([]byte) (int, error) })
	if !ok {
		return fmt.Errorf("invalid stdin reader")
	}
	out, ok := outWriter.(interface{ Write([]byte) (int, error) })
	if !ok {
		return fmt.Errorf("invalid stdout writer")
	}

	rw := newFramedRW(in, out)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		payload, err := rw.read(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		var msg rpcMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			_ = rw.writeJSON(ctx, rpcMessage{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &rpcError{Code: codeParseError, Message: "Parse error"},
			})
			continue
		}

		if msg.Method == "" {
			continue
		}

		resp, isNotification, perr := s.ProcessMessage(ctx, msg, false)
		if perr != nil {
			return perr
		}
		if isNotification || resp == nil {
			continue
		}
		if err := rw.writeJSON(ctx, *resp); err != nil {
			return err
		}
	}
}

func (s *Server) ProcessMessage(ctx context.Context, msg rpcMessage, httpTransport bool) (resp *rpcMessage, isNotification bool, err error) {
	if msg.Method == "" {
		return nil, true, nil
	}

	result, rerr := s.handle(ctx, msg.Method, msg.Params, httpTransport)

	if msg.ID == nil {
		return nil, true, nil
	}

	out := &rpcMessage{JSONRPC: "2.0", ID: msg.ID}
	if rerr != nil {
		out.Error = toRPCError(rerr)
	} else {
		out.Result = result
	}
	return out, false, nil
}

func (s *Server) handle(ctx context.Context, method string, params json.RawMessage, httpTransport bool) (any, error) {
	switch method {
	case "initialize":
		pv := negotiateProtocol(params)
		if httpTransport && pv == defaultProtocol {
			pv = protocolVersionHTTP
		}
		return map[string]any{
			"protocolVersion": pv,
			"serverInfo": map[string]any{
				"name":    version.ServerName,
				"version": version.MCP(),
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		}, nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		return listToolsPayload(s.runner.Registry.ListAll()), nil

	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, rpcErrf(codeInvalidParams, "bad params: %v", err)
		}
		if p.Name == "" {
			return nil, rpcErr(codeInvalidParams, "tool name required")
		}
		ctx, err := s.authorizeToolCall(ctx)
		if err != nil {
			return nil, err
		}
		return s.callTool(ctx, p.Name, p.Arguments)

	case "notifications/initialized":
		return nil, nil
	}

	return nil, rpcErrf(codeMethodNotFound, "unknown method: %s", method)
}

func (s *Server) authorizeToolCall(ctx context.Context) (context.Context, error) {
	if s.auth == nil || !s.auth.Config.Enabled {
		return ctx, nil
	}
	if sub, ok := auth.SubjectFromContext(ctx); ok {
		if err := s.auth.Enforcer.Enforce(sub, auth.PermEngageToolRun); err != nil {
			if errors.Is(err, auth.ErrForbidden) {
				return ctx, rpcErr(codeAuthError, "forbidden")
			}
			return ctx, rpcErr(codeAuthError, "unauthorized")
		}
		return ctx, nil
	}
	ctx, err := auth.AuthorizeEngageMCP(ctx, s.auth, "")
	if err != nil {
		if errors.Is(err, auth.ErrForbidden) {
			return ctx, rpcErr(codeAuthError, "forbidden")
		}
		return ctx, rpcErr(codeAuthError, "unauthorized")
	}
	return ctx, nil
}

func toRPCError(err error) *rpcError {
	var re *rpcError
	if errors.As(err, &re) {
		return re
	}
	if errors.Is(err, auth.ErrForbidden) {
		return &rpcError{Code: codeAuthError, Message: "forbidden"}
	}
	if errors.Is(err, auth.ErrUnauthorized) {
		return &rpcError{Code: codeAuthError, Message: "unauthorized"}
	}
	return &rpcError{Code: codeInternal, Message: err.Error()}
}

// CatalogCount exposes registry size for healthchecks.
func CatalogCount(reg *tools.Registry) int {
	if reg == nil {
		return 0
	}
	return reg.Count()
}
