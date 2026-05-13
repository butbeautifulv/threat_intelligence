package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/threat_intelligence/api/internal/components"
	"github.com/butbeautifulv/threat_intelligence/api/internal/config"
	"github.com/butbeautifulv/threat_intelligence/api/internal/transport/httpserver"
)

func main() {
	cfg := config.Load()
	logger := components.SetupLogger(cfg.Env)

	c, err := components.InitComponents(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Shutdown()

	mux := http.NewServeMux()
	httpserver.Register(mux, c.Graph)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shctx, cc := context.WithTimeout(context.Background(), 8*time.Second)
		defer cc()
		_ = srv.Shutdown(shctx)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	logger.Info("api listening", slog.String("addr", cfg.ListenAddr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
