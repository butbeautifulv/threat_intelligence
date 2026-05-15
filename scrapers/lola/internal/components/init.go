package components

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ingestpub"
	"lola/internal/config"
	lolanats "lola/internal/natspub"
	"lola/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	Scraper *usecase.ScraperUsecase
	NatsPub *ingestpub.JetStreamPublisher
}

func InitComponents(_ *config.Config, logger *slog.Logger) (*Components, error) {
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
		Scraper: scraper,
		NatsPub: pub,
	}, nil
}

func (c *Components) Shutdown() {
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
