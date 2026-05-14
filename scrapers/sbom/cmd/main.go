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
	neo4jstore "sbom/storage/neo4j"
)

func main() {
	sources := flag.String("sources", "", "comma-separated: osv, ghsa (default from SBOM_SOURCES)")
	maxCVE := flag.Int("max-cves", 0, "max CVEs (0 = use SBOM_MAX_CVES)")
	maxGHSA := flag.Int("max-ghsa", 0, "max GHSA (0 = use SBOM_MAX_GHSA)")
	minYear := flag.Int("ghsa-min-year", 0, "min year (0 = use SBOM_GHSA_MIN_YEAR)")
	ingestMode := flag.String("ingest-mode", "", "direct or nats (default INGEST_MODE env)")
	flag.Parse()

	ctx := context.Background()
	cfg := config.FromEnv()
	opt := usecase.Options{
		Sources:     cfg.Sources,
		MaxCVE:      cfg.MaxCVE,
		MaxGHSA:     cfg.MaxGHSA,
		GHSAMinYear: cfg.GHSAMinYear,
		IngestMode:  cfg.IngestMode,
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
	if strings.TrimSpace(*ingestMode) != "" {
		opt.IngestMode = strings.ToLower(strings.TrimSpace(*ingestMode))
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var (
		st  *neo4jstore.Store
		err error
	)
	if opt.IngestMode != config.IngestModeNATS {
		st, err = neo4jstore.New(ctx, neo4jstore.Config{
			URI:      cfg.Neo4jURI,
			Username: cfg.Neo4jUser,
			Password: cfg.Neo4jPass,
			Database: cfg.Neo4jDB,
		})
		if err != nil {
			log.Error("neo4j", slog.String("err", err.Error()))
			os.Exit(1)
		}
		defer st.Close(ctx)
		if err := st.EnsureSchema(ctx); err != nil {
			log.Error("schema", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}

	var pub *ingestpub.JetStreamPublisher
	if opt.IngestMode == config.IngestModeNATS {
		pub, err = ingestpub.ConnectJetStreamAndStream(cfg.NATSURL)
		if err != nil {
			log.Error("nats", slog.String("err", err.Error()))
			os.Exit(1)
		}
		defer pub.Close()
	}

	var cveSrc cvesource.Lister
	if opt.IngestMode == config.IngestModeNATS {
		cveSrc, err = cvesource.New(cfg.CVEListFile, cfg.CVEListURL)
		if err != nil {
			log.Error("cve list", slog.String("err", err.Error()))
			os.Exit(1)
		}
	} else {
		cveSrc = st
	}

	runner := usecase.NewRunner(log, st, cveSrc, pub, opt)
	if err := runner.Run(ctx); err != nil {
		log.Error("run", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("sbom scrape done")
}
