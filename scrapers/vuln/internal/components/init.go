package components

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"ingestpub"
	"vuln/internal/config"
	vulnnats "vuln/internal/natspub"
	"vuln/internal/repository"
	neo4jstore "vuln/internal/storage/neo4j"
	"vuln/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	VulnRepo   repository.VulnerabilityRepository
	Scraper    *usecase.ScraperUsecase
	Neo4jStore *neo4jstore.Store
	NatsPub    *ingestpub.JetStreamPublisher
}

func InitComponents(cfg *config.Config, logger *slog.Logger) (*Components, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("INGEST_MODE")))
	if mode == "nats" {
		natsURL := strings.TrimSpace(os.Getenv("NATS_URL"))
		if natsURL == "" {
			natsURL = "nats://localhost:4222"
		}
		subj := strings.TrimSpace(os.Getenv("VULN_NATS_SUBJECT"))
		if subj == "" {
			subj = "ingest.vuln.events"
		}
		pub, err := ingestpub.ConnectJetStreamAndStream(natsURL)
		if err != nil {
			return nil, err
		}
		repo := vulnnats.New(pub, subj)
		scraper := usecase.NewScraperUsecase(repo, logger, cfg.NVD.APIKey)
		return &Components{
			VulnRepo:   repo,
			Scraper:    scraper,
			Neo4jStore: nil,
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

	scraper := usecase.NewScraperUsecase(store, logger, cfg.NVD.APIKey)

	return &Components{
		VulnRepo:   store,
		Scraper:    scraper,
		Neo4jStore: store,
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
