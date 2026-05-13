package usecase

import (
	"log/slog"

	"mcp/internal/connector/neo4jconn"

	gq "github.com/butbeautifulv/threat_intelligence/graph/query"
)

// QueryUsecase wraps shared graph/query.Service for MCP (stdio tools).
type QueryUsecase struct {
	*gq.Service
	logger *slog.Logger
}

func NewQueryUsecase(conn *neo4jconn.Connector, logger *slog.Logger) *QueryUsecase {
	return &QueryUsecase{
		Service: gq.NewService(conn),
		logger:  logger,
	}
}
