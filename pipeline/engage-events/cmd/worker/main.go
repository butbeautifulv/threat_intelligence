package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connats "github.com/butbeautifulv/veil/pipeline/connector/nats"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	natsURL := env("NATS_URL", "nats://127.0.0.1:4222")
	filter := env("ENGAGE_EVENTS_FILTER", "engage.events.>")
	ingestRun := env("INGEST_SUBJECT", "ingest.engage.tool_run")
	ingestFinding := env("INGEST_FINDING_SUBJECT", "ingest.engage.finding")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger.Info("engage-events worker starting", slog.String("nats", natsURL), slog.String("filter", filter))
	if err := connats.RunEngageEventsConsumer(ctx, logger, natsURL, filter, ingestRun, ingestFinding); err != nil && ctx.Err() == nil {
		logger.Error("consumer stopped", slog.Any("err", err))
		os.Exit(1)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
