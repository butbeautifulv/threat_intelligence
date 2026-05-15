package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/butbeautifulv/threat_intelligence/scrape/factory"

	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/coderules/scrapesource"
	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/ds/scrapesource"
	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/lola/scrapesource"
	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/nuclei/scrapesource"
	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/sbom/scrapesource"
	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/ti/scrapesource"
	_ "github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln/scrapesource"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	names := factory.ParseSourceNames(os.Getenv("SCRAPE_SOURCES"))
	logger.Info("scrape_worker starting", slog.Any("sources", names))

	if err := factory.Run(context.Background(), factory.RunOptions{
		SourceNames: names,
		Log:         logger,
	}); err != nil {
		log.Fatal(err)
	}
}
