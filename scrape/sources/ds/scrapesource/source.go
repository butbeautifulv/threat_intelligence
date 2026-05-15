// Package scrapesource registers the ds scrape source with scrape/factory.
package scrapesource

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/scrape/factory"
	dsscrapepub "github.com/butbeautifulv/threat_intelligence/scrape/sources/ds/internal/scrapepub"
	"github.com/butbeautifulv/threat_intelligence/scrape/sources/ds/internal/usecase"
)

func init() {
	factory.Register("ds", func() factory.Source { return &Source{} })
}

// Source scrapes Sigma, YARA, Atomic Red Team, and Caldera via GitHub.
type Source struct{}

func (s *Source) Name() string { return "ds" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	pub, err := deps.Publisher("ds")
	if err != nil {
		return err
	}
	repo := dsscrapepub.NewFromRaw(pub)
	ing := usecase.NewIngestor(repo, deps.Log, cacheDir(), deps.Feeds, deps.Ledger)
	return ing.Run(ctx)
}

func cacheDir() string {
	if v := strings.TrimSpace(os.Getenv("DS_CACHE_DIR")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("SCRAPE_CACHE_DIR")); v != "" {
		return v
	}
	return filepath.Join(".", "data", "cache")
}
