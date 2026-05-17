package tooldispatch

import (
	"fmt"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ArgsFromRequest converts HTTP tool run body to dispatch argument map.
func ArgsFromRequest(req contract.ToolRunRequest) map[string]any {
	args := make(map[string]any)
	if req.Target != "" {
		args["target"] = req.Target
	}
	if req.AdditionalArgs != "" {
		args["additional_args"] = req.AdditionalArgs
	}
	for k, v := range req.Parameters {
		args[k] = v
	}
	return args
}

// RequestFromArgs converts MCP-style arguments to ToolRunRequest.
func RequestFromArgs(args map[string]any) contract.ToolRunRequest {
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
