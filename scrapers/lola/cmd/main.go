package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"lola/internal/components"
	"lola/internal/config"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	logger := components.SetupLogger(cfg.Env)

	c, err := components.InitComponents(cfg, logger)
	if err != nil {
		logger.Error("bad configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer c.Shutdown()

	eg, ctx := errgroup.WithContext(context.Background())

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)

	eg.Go(func() error {
		logger.Info("LOLA ingest started")
		return c.Scraper.Run(ctx)
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sigQuit:
			logger.Info("Captured signal", slog.String("signal", s.String()))
			return fmt.Errorf("captured signal: %v", s)
		}
	})

	err = eg.Wait()
	logger.Info("shutdown", slog.String("error", fmt.Sprint(err)))
}
