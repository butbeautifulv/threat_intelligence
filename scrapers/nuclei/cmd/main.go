package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"ingestpub"
	"nuclei/internal/config"
	"nuclei/internal/usecase"
)

func main() {
	max := flag.Int("max", 0, "max templates (0 = NUCLEI_MAX)")
	years := flag.String("years", "", "comma-separated CVE year folders (default NUCLEI_YEARS)")
	flag.Parse()

	ctx := context.Background()
	cfg := config.FromEnv()
	opt := usecase.Options{
		MaxTemplates: cfg.MaxTemplates,
		YearsCSV:     cfg.YearsCSV,
		NATSURL:      cfg.NATSURL,
		NATSSubject:  cfg.NATSSubject,
	}
	if *max > 0 {
		opt.MaxTemplates = *max
	}
	if strings.TrimSpace(*years) != "" {
		opt.YearsCSV = strings.TrimSpace(*years)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	pub, err := ingestpub.ConnectJetStreamAndStream(cfg.NATSURL)
	if err != nil {
		log.Error("nats", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pub.Close()

	runner := usecase.NewRunner(log, pub, opt)
	if err := runner.Run(ctx); err != nil {
		log.Error("run", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
