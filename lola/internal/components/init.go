package components

import (
	"log/slog"
	"os"
	"vuln/internal/config"
	"vuln/internal/repository"
	"vuln/internal/storage/mongo"
	"vuln/internal/usecase"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	MongoClient *mongo.ClientWrapper
	VulnRepo    repository.VulnerabilityRepository
	Scraper     *usecase.ScraperUsecase
}

func InitComponents(cfg *config.Config, logger *slog.Logger) (*Components, error) {
	// 1. Mongo client (инфраструктура)
	mongoClient, err := mongo.NewMongoClient(cfg.MongoConfig, logger)
	if err != nil {
		return nil, err
	}

	// 2. Mongo repository (адаптер)
	vulnRepo := mongo.NewVulnerabilityRepository(
		mongoClient.Database(cfg.MongoConfig.Database),
	)

	// 3. Usecase (бизнес-логика)
	scraper := usecase.NewScraperUsecase(vulnRepo, logger)

	return &Components{
		MongoClient: mongoClient,
		VulnRepo:    vulnRepo,
		Scraper:     scraper,
	}, nil
}

func (c *Components) Shutdown() {
	if c.MongoClient != nil {
		c.MongoClient.Close()
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
