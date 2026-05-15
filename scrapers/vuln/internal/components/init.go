package components

import (
	"log/slog"
	"os"
	"strings"

	"ingestpub"
	"vuln/internal/config"
	vulnnats "vuln/internal/natspub"
	"vuln/internal/repository"
	"vuln/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	VulnRepo repository.VulnerabilityRepository
	Scraper  *usecase.ScraperUsecase
	NatsPub  *ingestpub.JetStreamPublisher
}

func InitComponents(cfg *config.Config, logger *slog.Logger) (*Components, error) {
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
		VulnRepo: repo,
		Scraper:  scraper,
		NatsPub:  pub,
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
