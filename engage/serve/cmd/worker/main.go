package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(0)
	}
	cfg := config.LoadAPI()
	logger := components.SetupLogger(cfg.Env)
	api, err := components.InitAPI(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	queue := jobuc.NewQueue(api.Tools, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	log.Println("engage worker: async job queue ready (enqueue via API in future releases)")
	if err := queue.RunWorker(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
