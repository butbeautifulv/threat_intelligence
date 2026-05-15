package components

import (
	"context"
	"log/slog"
	"os"

	"github.com/butbeautifulv/threat_intelligence/graph/api/internal/config"
	neo4jstore "github.com/butbeautifulv/threat_intelligence/graph/api/internal/storage/neo4j"
	"github.com/butbeautifulv/threat_intelligence/graph/api/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	Neo4jStore *neo4jstore.Store
	Graph      *usecase.GraphUsecase
}

func InitComponents(cfg *config.Config) (*Components, error) {
	store, err := neo4jstore.New(context.Background(), neo4jstore.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
		Database: cfg.Neo4j.Database,
	})
	if err != nil {
		return nil, err
	}

	uc := usecase.NewGraphUsecase(store)

	return &Components{
		Neo4jStore: store,
		Graph:      uc,
	}, nil
}

func (c *Components) Shutdown() {
	if c.Neo4jStore != nil {
		_ = c.Neo4jStore.Close(context.Background())
	}
}

func SetupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
}
