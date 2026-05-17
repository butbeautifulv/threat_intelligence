package mcpserver

import (
	"context"
	"errors"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tooldispatch"
)

func (s *Server) callTool(ctx context.Context, name string, args map[string]any) (any, error) {
	body, err := s.dispatch.Dispatch(ctx, "", name, args)
	if err != nil {
		var de *tooldispatch.DispatchError
		if errors.As(err, &de) {
			if de.NotFound {
				return nil, rpcErrf(codeMethodNotFound, "%s", de.Message)
			}
			return nil, rpcErrf(codeToolError, "%s", de.Message)
		}
		return nil, err
	}
	return toolTextResult(body)
}
