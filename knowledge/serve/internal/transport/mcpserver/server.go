package mcpserver

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/mcp"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/version"
)

type Server struct {
	uc     *usecase.ReadUsecase
	auth   *auth.Stack
	logger *slog.Logger
}

func NewServer(uc *usecase.ReadUsecase, stack *auth.Stack, logger *slog.Logger) *Server {
	return &Server{uc: uc, auth: stack, logger: logger}
}

func (s *Server) Run(ctx context.Context, inReader any, outWriter any) error {
	return mcp.RunStdio(ctx, s, inReader, outWriter)
}

// ProcessMessage handles one JSON-RPC message. httpTransport selects protocol version and transport quirks.
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
		return listToolsPayload(), nil

	case "tools/call":
		p, err := mcp.ParseToolCallParams(params)
		if err != nil {
			return nil, err
		}
		ctx, err = mcp.AuthorizeToolCall(ctx, s.auth, auth.PermGraphRead, auth.AuthorizeMCP)
		if err != nil {
			return nil, err
		}
		return s.callTool(ctx, p.Name, p.Arguments)

	case "notifications/initialized":
		return nil, nil
	}

	return nil, rpcErrf(codeMethodNotFound, "unknown method: %s", method)
}
