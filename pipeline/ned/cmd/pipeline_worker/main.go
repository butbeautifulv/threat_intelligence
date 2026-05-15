package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/butbeautifulv/veil/pipeline/ned/internal/components"
)

func main() {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := components.Run(rootCtx, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("shutdown complete")
}
