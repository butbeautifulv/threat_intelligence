package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	neo4jstore "ds/internal/storage/neo4j"
	"ds/internal/usecase"
)

func main() {
	ctx := context.Background()

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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cache := envOr("DS_CACHE_DIR", filepath.Join(".", "data", "cache"))
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
