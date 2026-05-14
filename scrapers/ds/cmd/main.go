package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ingestpub"
	dsnats "ds/internal/natspub"
	neo4jstore "ds/internal/storage/neo4j"
	"ds/internal/usecase"
)

func main() {
	ctx := context.Background()
	mode := strings.ToLower(strings.TrimSpace(envOr("INGEST_MODE", "direct")))

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cache := envOr("DS_CACHE_DIR", filepath.Join(".", "data", "cache"))

	if mode == "nats" {
		natsURL := strings.TrimSpace(envOr("NATS_URL", "nats://localhost:4222"))
		subj := strings.TrimSpace(envOr("DS_NATS_SUBJECT", "ingest.ds.events"))
		pub, err := ingestpub.ConnectJetStreamAndStream(natsURL)
		if err != nil {
			log.Fatal(err)
		}
		defer pub.Close()
		repo := dsnats.New(pub, subj)
		ing := usecase.NewIngestor(repo, logger, cache)
		if err := ing.Run(ctx); err != nil {
			log.Fatal(err)
		}
		return
	}

	store, err := neo4jstore.New(ctx, neo4jstore.Config{
		URI:      envOr("NEO4J_URI", "neo4j://localhost:7687"),
		Username: envOr("NEO4J_USER", "neo4j"),
		Password: envOr("NEO4J_PASS", "neo4jpassword"),
		Database: envOr("NEO4J_DB", "neo4j"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = store.Close(cctx)
	}()

	ing := usecase.NewIngestor(store, logger, cache)
	if err := ing.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
