package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/butbeautifulv/threat_intelligence/graph/ingest/internal/components"
	"github.com/butbeautifulv/threat_intelligence/graph/ingest/internal/config"
	graphnats "github.com/butbeautifulv/threat_intelligence/graph/ingest/internal/connector/nats"
	ingestloop "github.com/butbeautifulv/threat_intelligence/graph/ingest/internal/ingest"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

func main() {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(rootCtx, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("shutdown complete")
}

func run(rootCtx context.Context, log *slog.Logger) error {
	cfg := config.Load()

	rt, err := components.Init(rootCtx, cfg, log)
	if err != nil {
		return err
	}
	defer rt.Shutdown(rootCtx)

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	if err := graphnats.EnsureIngestStream(js); err != nil {
		return err
	}

	sub, err := js.PullSubscribe(cfg.Subject, cfg.Durable, nats.BindStream(cfg.Stream))
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	log.Info("ingest-worker started",
		slog.String("nats", cfg.NATSURL),
		slog.String("stream", cfg.Stream),
		slog.String("durable", cfg.Durable),
		slog.String("subject", cfg.Subject),
	)

	eg, ctx := errgroup.WithContext(rootCtx)
	eg.Go(func() error {
		return ingestloop.RunPullLoop(ctx, log, sub, cfg.Batch, cfg.MaxWait, rt)
	})
	return eg.Wait()
}
