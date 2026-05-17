package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"

	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/coderules/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ds/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/lola/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/nuclei/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/vuln/scrapesource"
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
