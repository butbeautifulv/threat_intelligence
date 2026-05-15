package components

import (
	"context"
	"log/slog"

	"github.com/butbeautifulv/veil/graph/serve/internal/config"
	"github.com/butbeautifulv/veil/graph/serve/internal/connector"
	"github.com/butbeautifulv/veil/graph/serve/internal/transport/mcpserver"
	"github.com/butbeautifulv/veil/graph/serve/internal/usecase"
)

type MCPComponents struct {
	Neo4j     *connector.ReadConnector
	Query     *usecase.QueryUsecase
	MCPServer *mcpserver.Server
}

func InitMCP(cfg *config.Config, logger *slog.Logger) (*MCPComponents, error) {
	conn, err := connector.NewRead(context.Background(), connector.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
		Database: cfg.Neo4j.Database,
	})
	if err != nil {
		return nil, err
	}
	uc := usecase.NewQueryUsecase(conn, logger)
	return &MCPComponents{
		Neo4j:     conn,
		Query:     uc,
		MCPServer: mcpserver.NewServer(uc, logger),
	}, nil
}

func (c *MCPComponents) Shutdown() {
	if c.Neo4j != nil {
		_ = c.Neo4j.Close(context.Background())
	}
}
