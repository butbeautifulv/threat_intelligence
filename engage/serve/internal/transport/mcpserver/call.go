package mcpserver

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

func (s *Server) callTool(ctx context.Context, name string, args map[string]any) (any, error) {
	target := argTarget(args)
	subject := ""
	if sub, ok := auth.SubjectFromContext(ctx); ok {
		subject = sub.Sub
	}
	if out, ok, err := s.tryPlaybookByName(ctx, subject, name, target, argBool(args, "async")); ok {
		return out, err
	}
	if out, ok, err := s.tryAgentTool(ctx, name, args); ok {
		return out, err
	}
	spec, ok := s.runner.Registry.Get(name)
	if !ok {
		return nil, rpcErrf(codeMethodNotFound, "unknown tool: %s", name)
	}
	if IsIntelBridgeTool(name, spec) {
		return s.callIntelBridge(ctx, name, spec, args)
	}
	if IsBridgeWorkflowBinary(spec.Binary) {
		if out, ok, err := s.tryBridgeWorkflowTool(ctx, name, args); ok {
			return out, err
		}
	}
	res := s.runner.Run(ctx, subject, name, argsToRequest(args))
	if !res.Success && res.Error != "" {
		return nil, rpcErrf(codeToolError, "%s", res.Error)
	}
	return toolTextResult(res)
}

func argsToRequest(args map[string]any) contract.ToolRunRequest {
	req := contract.ToolRunRequest{Parameters: make(map[string]string)}
	if args == nil {
		return req
	}
	for k, v := range args {
		switch k {
		case "target", "url", "domain", "host":
			if req.Target == "" {
				req.Target = fmt.Sprint(v)
			}
			req.Parameters[k] = fmt.Sprint(v)
		case "additional_args":
			req.AdditionalArgs = fmt.Sprint(v)
		default:
			req.Parameters[k] = fmt.Sprint(v)
		}
	}
	if req.Target == "" {
		if t, ok := req.Parameters["target"]; ok {
			req.Target = t
		}
	}
	return req
}
