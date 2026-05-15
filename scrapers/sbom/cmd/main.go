package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"ingestpub"
	"sbom/internal/config"
	"sbom/internal/cvesource"
	"sbom/internal/usecase"
)

func main() {
	sources := flag.String("sources", "", "comma-separated: osv, ghsa (default from SBOM_SOURCES)")
	maxCVE := flag.Int("max-cves", 0, "max CVEs (0 = use SBOM_MAX_CVES)")
	maxGHSA := flag.Int("max-ghsa", 0, "max GHSA (0 = use SBOM_MAX_GHSA)")
	minYear := flag.Int("ghsa-min-year", 0, "min year (0 = use SBOM_GHSA_MIN_YEAR)")
	flag.Parse()

	ctx := context.Background()
	cfg := config.FromEnv()
	opt := usecase.Options{
		Sources:     cfg.Sources,
		MaxCVE:      cfg.MaxCVE,
		MaxGHSA:     cfg.MaxGHSA,
		GHSAMinYear: cfg.GHSAMinYear,
		NATSURL:     cfg.NATSURL,
		NATSSubject: cfg.NATSSubject,
	}
	if *sources != "" {
		var p []string
		for _, s := range strings.Split(*sources, ",") {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				p = append(p, s)
			}
		}
		if len(p) > 0 {
			opt.Sources = p
		}
	}
	if *maxCVE > 0 {
		opt.MaxCVE = *maxCVE
	}
	if *maxGHSA > 0 {
		opt.MaxGHSA = *maxGHSA
	}
	if *minYear > 0 {
		opt.GHSAMinYear = *minYear
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cveSrc, err := cvesource.New(cfg.CVEListFile, cfg.CVEListURL)
	if err != nil {
		log.Error("cve list", slog.String("err", err.Error()))
		os.Exit(1)
	}

	pub, err := ingestpub.ConnectJetStreamAndStream(cfg.NATSURL)
	if err != nil {
		log.Error("nats", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pub.Close()

	runner := usecase.NewRunner(log, cveSrc, pub, opt)
	if err := runner.Run(ctx); err != nil {
		log.Error("run", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("sbom scrape done")
}
