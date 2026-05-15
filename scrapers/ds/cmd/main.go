package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ingestpub"
	dsnats "ds/internal/natspub"
	"ds/internal/usecase"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cache := envOr("DS_CACHE_DIR", filepath.Join(".", "data", "cache"))

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
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
