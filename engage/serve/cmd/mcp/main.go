package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/engage/serve/internal/transport/mcpserver"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	cfg := config.LoadMCP()
	logger := components.SetupMCPLogger(cfg.Env)

	if err := config.ValidateSecurity(cfg.Security, cfg.Auth.Enabled); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigQuit
		cancel()
	}()

	c, err := components.InitMCP(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := c.ListenMCPHTTP(ctx, cfg, logger); err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		os.Exit(0)
	}()

	if err := c.MCPServer.Run(ctx, os.Stdin, os.Stdout); err != nil {
		logger.Error("mcp server stopped", slog.Any("err", err))
		time.Sleep(200 * time.Millisecond)
		os.Exit(1)
	}
}

func runHealthcheck() int {
	cfg := config.LoadMCP()
	c, err := components.InitMCP(cfg, components.SetupMCPLogger(cfg.Env))
	if err != nil {
		return 1
	}
	if mcpserver.CatalogCount(c.API.Registry) < 1 {
		return 1
	}
	return 0
}
