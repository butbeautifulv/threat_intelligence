package ports

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ToolDispatcher routes tool execution for HTTP and MCP surfaces.
type ToolDispatcher interface {
	Dispatch(ctx context.Context, subject, name string, args map[string]any) (any, error)
	DispatchRequest(ctx context.Context, subject, name string, req contract.ToolRunRequest) (any, error)
}
