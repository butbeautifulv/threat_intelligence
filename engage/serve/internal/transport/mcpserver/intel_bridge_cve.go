package mcpserver

import "context"

type cveBridgeHandler func(ctx context.Context, s *Server, args map[string]any) (any, error)

var cveBridgeHandlers = map[string]cveBridgeHandler{
	"monitor_cve_feeds": func(ctx context.Context, s *Server, args map[string]any) (any, error) {
		return toolJSONResult(s.cve.MonitorFromBody(ctx, args))
	},
	"generate_exploit_from_cve": func(ctx context.Context, s *Server, args map[string]any) (any, error) {
		return toolJSONResult(s.cve.GenerateExploitFromCVE(ctx, args))
	},
}

func (s *Server) callCVEBridge(ctx context.Context, name string, args map[string]any) (any, error) {
	if s.cve == nil {
		return nil, rpcErrf(codeToolError, "CVE service not configured")
	}
	if h, ok := cveBridgeHandlers[name]; ok {
		return h(ctx, s, args)
	}
	return nil, rpcErrf(codeToolError, "unknown CVE tool %q", name)
}
