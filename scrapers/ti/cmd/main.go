package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ti/internal/feeds"
	"ti/internal/ingest"
	neo4jstore "ti/internal/storage/neo4j"
	"ti/internal/usecase"
)

func main() {
	var (
		input = flag.String("input", "", "optional path to TI JSONL file")
		feedsFlag = flag.String("feeds", "", "comma-separated public feeds: kev,pt,urlhaus,threatfox,malwarebazaar,feodo,openphish")

		neo4jURI  = flag.String("neo4j-uri", envOr("NEO4J_URI", "neo4j://localhost:7687"), "neo4j uri")
		neo4jUser = flag.String("neo4j-user", envOr("NEO4J_USER", "neo4j"), "neo4j username")
		neo4jPass = flag.String("neo4j-pass", envOr("NEO4J_PASS", "neo4jpassword"), "neo4j password")
		neo4jDB   = flag.String("neo4j-db", envOr("NEO4J_DB", "neo4j"), "neo4j database")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigQuit
		cancel()
	}()

	store, err := neo4jstore.New(ctx, neo4jstore.Config{
		URI:      *neo4jURI,
		Username: *neo4jUser,
		Password: *neo4jPass,
		Database: *neo4jDB,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cctx, ccancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer ccancel()
		_ = store.Close(cctx)
	}()

	if *feedsFlag != "" {
		kinds := strings.Split(*feedsFlag, ",")
		runner := feeds.NewRunner(store, logger)
		if err := runner.Run(ctx, kinds); err != nil {
			log.Fatal(err)
		}
	}

	if *input != "" {
		stream, closeFn, err := ingest.NewStreamFromFile(*input)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = closeFn() }()

		uc := usecase.NewIngestor(store, logger)
		if err := uc.IngestJSONL(ctx, stream); err != nil {
			log.Fatal(err)
		}
	}

	if *input == "" && *feedsFlag == "" {
		logger.Info("nothing to do: pass --feeds kev,urlhaus,threatfox,... and/or --input path/to.jsonl")
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
