package components

import (
	"context"
	"log/slog"
	"os"

	"mcp/internal/config"
	"mcp/internal/connector/neo4jconn"
	"mcp/internal/transport/mcpserver"
	"mcp/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	Neo4j      *neo4jconn.Connector
	Query      *usecase.QueryUsecase
	MCPServer  *mcpserver.Server
}

func InitComponents(cfg *config.Config, logger *slog.Logger) (*Components, error) {
	conn, err := neo4jconn.New(context.Background(), neo4jconn.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
		Database: cfg.Neo4j.Database,
	})
	if err != nil {
		return nil, err
	}

	uc := usecase.NewQueryUsecase(conn, logger)
	srv := mcpserver.NewServer(uc, logger)

	return &Components{
		Neo4j:     conn,
		Query:     uc,
		MCPServer: srv,
	}, nil
}

func (c *Components) Shutdown() {
	if c.Neo4j != nil {
		_ = c.Neo4j.Close(context.Background())
	}
}

func SetupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

