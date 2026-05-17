package mcpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/engage/serve/internal/version"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

func HTTPHandler(s *Server, cfg config.MCPHTTPConfig) http.Handler {
	path := cfg.Path
	if path == "" {
		path = "/mcp"
	}
	return mcp.HTTPHandler(s, mcp.HTTPConfig{
		Path:    path,
		Service: version.ServerName,
		HealthExtra: map[string]any{
			"tools": CatalogCount(s.runner.Registry),
		},
		Logger: s.logger,
	})
}
