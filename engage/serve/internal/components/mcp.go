package components

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	authmw "github.com/butbeautifulv/veil/engage/serve/internal/auth/middleware"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/engage/serve/internal/transport/mcpserver"
	"github.com/butbeautifulv/veil/engage/serve/internal/transport/securityhttp"
)

type MCPComponents struct {
	API       *APIComponents
	MCPServer *mcpserver.Server
}

func InitMCP(cfg *config.Config, logger *slog.Logger) (*MCPComponents, error) {
	api, err := InitAPI(cfg, logger)
	if err != nil {
		return nil, err
	}
	return &MCPComponents{
		API:       api,
		MCPServer: mcpserver.NewServerWithDispatch(api.ToolDispatch, api.Tools, api.Auth, logger),
	}, nil
}

func (c *MCPComponents) MCPHTTPHandler(cfg *config.Config) http.Handler {
	h := mcpserver.HTTPHandler(c.MCPServer, cfg.MCPHTTP)
	var inner http.Handler = h
	if c.API.Auth != nil && c.API.Auth.Config.Enabled && cfg.Security.MCPHTTPAuthStrict {
		inner = authmw.Auth(c.API.Auth, true, cfg.Security, h)
	}
	limit := cfg.Security.MCPBodyLimit
	if limit <= 0 {
		limit = 8 << 20
	}
	return securityhttp.Harden(cfg.Security, limit, inner)
}

func (c *MCPComponents) ListenMCPHTTP(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*http.Server, error) {
	if !cfg.MCPHTTP.Enabled {
		return nil, nil
	}
	addr := cfg.MCPHTTP.Listen
	if cfg.MCPHTTP.BindLocal {
		addr = "127.0.0.1" + addr
	}
	rh, rt, wt, idle := securityhttp.HTTPServerTimeouts()
	srv := &http.Server{
		Addr:              addr,
		Handler:           c.MCPHTTPHandler(cfg),
		ReadHeaderTimeout: time.Duration(rh) * time.Second,
		ReadTimeout:       time.Duration(rt) * time.Second,
		WriteTimeout:      time.Duration(wt) * time.Second,
		IdleTimeout:       time.Duration(idle) * time.Second,
	}
	go func() {
		logger.Info("engage mcp http listening", slog.String("addr", addr), slog.String("path", cfg.MCPHTTP.Path))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("engage mcp http stopped", slog.Any("err", err))
		}
	}()
	go func() {
		<-ctx.Done()
		shctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		_ = srv.Shutdown(shctx)
	}()
	return srv, nil
}
