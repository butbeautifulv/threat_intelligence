package components

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ingestpub"
	"lola/internal/config"
	lolanats "lola/internal/natspub"
	neo4jstore "lola/internal/storage/neo4j"
	"lola/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	Neo4jStore *neo4jstore.Store
	Scraper    *usecase.ScraperUsecase
	NatsPub    *ingestpub.JetStreamPublisher
}

func InitComponents(cfg *config.Config, logger *slog.Logger) (*Components, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("INGEST_MODE")))
	if mode == "nats" {
		natsURL := strings.TrimSpace(os.Getenv("NATS_URL"))
		if natsURL == "" {
			natsURL = "nats://localhost:4222"
		}
		subj := strings.TrimSpace(os.Getenv("LOLA_NATS_SUBJECT"))
		if subj == "" {
			subj = "ingest.lola.events"
		}
		pub, err := ingestpub.ConnectJetStreamAndStream(natsURL)
		if err != nil {
			return nil, err
		}
		repo := lolanats.New(pub, subj)
		cache := os.Getenv("LOLA_CACHE_DIR")
		if cache == "" {
			wd, err := os.Getwd()
			if err != nil {
				wd = "."
			}
			cache = filepath.Join(wd, "data", "cache")
		}
		scraper := usecase.NewScraperUsecase(repo, logger, cache)
		return &Components{
			Neo4jStore: nil,
			Scraper:    scraper,
			NatsPub:    pub,
		}, nil
	}

	store, err := neo4jstore.New(context.Background(), neo4jstore.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
		Database: cfg.Neo4j.Database,
	})
	if err != nil {
		return nil, err
	}
	if err := store.EnsureSchema(context.Background()); err != nil {
		_ = store.Close(context.Background())
		return nil, err
	}

	cache := os.Getenv("LOLA_CACHE_DIR")
	if cache == "" {
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		cache = filepath.Join(wd, "data", "cache")
	}
	scraper := usecase.NewScraperUsecase(store, logger, cache)

	return &Components{
		Neo4jStore: store,
		Scraper:    scraper,
		NatsPub:    nil,
	}, nil
}

func (c *Components) Shutdown() {
	if c.Neo4jStore != nil {
		_ = c.Neo4jStore.Close(context.Background())
	}
	if c.NatsPub != nil {
		c.NatsPub.Close()
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
