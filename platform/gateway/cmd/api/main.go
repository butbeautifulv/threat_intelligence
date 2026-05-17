package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
	"github.com/butbeautifulv/veil/platform/gateway/internal/httpserver"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	cfg := config.Load()
	mux := http.NewServeMux()
	httpserver.Register(mux, cfg, nil)

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()
	go func() {
		<-ctx.Done()
		shctx, cc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cc()
		_ = srv.Shutdown(shctx)
	}()

	log.Printf("veil-api gateway listening on %s (graph=%s engage=%s)",
		cfg.ListenAddr, cfg.GraphAPIURL, cfg.EngageAPIURL)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func runHealthcheck() int {
	cfg := config.Load()
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1"+cfg.ListenAddr+"/health", nil)
	if err != nil {
		return 1
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
