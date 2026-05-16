package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	authmw "github.com/butbeautifulv/veil/engage/serve/internal/auth/middleware"
	"github.com/butbeautifulv/veil/engage/serve/internal/transport/httpserver"
	"github.com/butbeautifulv/veil/engage/serve/internal/transport/securityhttp"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(0)
	}

	cfg := config.LoadAPI()
	logger := components.SetupLogger(cfg.Env)

	c, err := components.InitAPI(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	httpserver.Register(mux, c)
	handler := securityhttp.Harden(cfg.Security, cfg.Security.APIBodyLimit,
		authmw.Auth(c.Auth, false, cfg.Security, mux))

	rh, rt, wt, idle := securityhttp.HTTPServerTimeouts()
	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(rh) * time.Second,
		ReadTimeout:       time.Duration(rt) * time.Second,
		WriteTimeout:      time.Duration(wt) * time.Second,
		IdleTimeout:       time.Duration(idle) * time.Second,
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

	log.Printf("veil-engage api listening on %s (%d tools)", cfg.ListenAddr, c.Registry.Count())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
