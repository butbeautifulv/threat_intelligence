// Package scrapesource registers the vuln scrape source with scrape/factory.
package scrapesource

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/scrape/factory"
	vulnscrapepub "github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln/internal/scrapepub"
	"github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln/internal/config"
	"github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln/internal/usecase"
)

func init() {
	factory.Register("vuln", func() factory.Source { return &Source{} })
}

// Source scrapes NVD, Metasploit, Exploit-DB, and optional Vulners.
type Source struct{}

func (s *Source) Name() string { return "vuln" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	pub, err := deps.Publisher("vuln")
	if err != nil {
		return err
	}
	repo := vulnscrapepub.NewFromRaw(pub)
	scraper := usecase.NewScraperUsecase(repo, deps.Log, cfg.NVD.APIKey, deps.Feeds, deps.Ledger)
	return scraper.Run(ctx)
}
