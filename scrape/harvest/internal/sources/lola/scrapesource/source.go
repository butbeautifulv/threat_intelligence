// Package scrapesource registers the lola scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/factory"
	lolascrapepub "github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/sources/lola/internal/scrapepub"
	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/sources/lola/internal/usecase"
)

func init() {
	factory.Register("lola", func() factory.Source { return &Source{} })
}

// Source scrapes LOLBAS, GTFOBins, LOFTS, and MITRE ATT&CK STIX.
type Source struct{}

func (s *Source) Name() string { return "lola" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	pub, err := deps.Publisher("lola")
	if err != nil {
		return err
	}
	repo := lolascrapepub.NewFromRaw(pub)
	scraper := usecase.NewScraperUsecase(repo, deps.Log, cacheDir(), deps.Feeds, deps.Ledger)
	return scraper.Run(ctx)
}

func cacheDir() string {
	if v := strings.TrimSpace(os.Getenv("LOLA_CACHE_DIR")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("SCRAPE_CACHE_DIR")); v != "" {
		return v
	}
	return filepath.Join(".", "data", "cache")
}
