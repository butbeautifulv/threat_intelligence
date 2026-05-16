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
	switch cfg.JobsMode {
	case jobuc.ModeFile, jobuc.ModeRedis, jobuc.ModeNats:
	default:
		cfg.JobsMode = jobuc.ModeFile
	}
	logger := components.SetupLogger(cfg.Env)
	api, err := components.InitAPI(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	log.Printf("engage worker: mode=%s", cfg.JobsMode)
	if err := api.Jobs.RunWorker(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
