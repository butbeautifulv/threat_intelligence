package usecase

import (
	"log/slog"

	"github.com/butbeautifulv/threat_intelligence/graph/connector/query"
	"github.com/butbeautifulv/threat_intelligence/graph/serve/internal/connector"
)

// QueryUsecase wraps graph/connector/query for MCP stdio tools.
type QueryUsecase struct {
	*query.Service
	logger *slog.Logger
}

func NewQueryUsecase(conn *connector.ReadConnector, logger *slog.Logger) *QueryUsecase {
	return &QueryUsecase{
		Service: query.NewService(conn),
		logger:  logger,
	}
}
