package scrape

import (
	"context"
	"log/slog"
)

type ScraperUsecase struct {
	repo   repository.VulnerabilityRepository
	logger *slog.Logger
}

func (u *ScraperUsecase) ScrapeNVD(ctx context.Context) error {
	data, err := downloadNVDFeed()
	if err != nil {
		return err
	}

	vulns := parseNVD(data)

	for _, v := range vulns {
		if err := u.repo.Upsert(ctx, &v); err != nil {
			return err
		}
	}

	return nil
}
