package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"ingestpub"
	"nuclei/internal/config"
	"nuclei/internal/usecase"
	neo4jstore "nuclei/storage/neo4j"
)

func main() {
	max := flag.Int("max", 0, "max templates (0 = NUCLEI_MAX)")
	years := flag.String("years", "", "comma-separated CVE year folders (default NUCLEI_YEARS)")
	ingestMode := flag.String("ingest-mode", "", "direct or nats (default INGEST_MODE)")
	flag.Parse()

	ctx := context.Background()
	cfg := config.FromEnv()
	opt := usecase.Options{
		MaxTemplates: cfg.MaxTemplates,
		YearsCSV:     cfg.YearsCSV,
		IngestMode:   cfg.IngestMode,
		NATSURL:      cfg.NATSURL,
		NATSSubject:  cfg.NATSSubject,
	}
	if *max > 0 {
		opt.MaxTemplates = *max
	}
	if strings.TrimSpace(*years) != "" {
		opt.YearsCSV = strings.TrimSpace(*years)
	}
	if strings.TrimSpace(*ingestMode) != "" {
		opt.IngestMode = strings.ToLower(strings.TrimSpace(*ingestMode))
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var st *neo4jstore.Store
	var err error
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

	runner := usecase.NewRunner(log, st, pub, opt)
	if err := runner.Run(ctx); err != nil {
		log.Error("run", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
