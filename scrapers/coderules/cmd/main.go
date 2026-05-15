package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"coderules/internal/config"
	"coderules/internal/usecase"
	"ingestpub"
)

func main() {
	sources := flag.String("sources", "", "comma-separated: cwe, semgrep, codeql (default CODERULES_SOURCES)")
	maxCWE := flag.Int("max-cwe", 0, "max CWE (0 = CODERULES_MAX_CWE)")
	maxSemgrep := flag.Int("max-semgrep", 0, "max Semgrep (0 = CODERULES_MAX_SEMGREP)")
	maxCodeQL := flag.Int("max-codeql", 0, "max CodeQL (0 = CODERULES_MAX_CODEQL)")
	flag.Parse()

	ctx := context.Background()
	cfg := config.FromEnv()
	opt := usecase.Options{
		Sources:     cfg.Sources,
		MaxCWE:      cfg.MaxCWE,
		MaxSemgrep:  cfg.MaxSemgrep,
		MaxCodeQL:   cfg.MaxCodeQL,
		NATSURL:     cfg.NATSURL,
		NATSSubject: cfg.NATSSubject,
	}
	if *sources != "" {
		var p []string
		for _, s := range strings.Split(*sources, ",") {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				p = append(p, s)
			}
		}
		if len(p) > 0 {
			opt.Sources = p
		}
	}
	if *maxCWE > 0 {
		opt.MaxCWE = *maxCWE
	}
	if *maxSemgrep > 0 {
		opt.MaxSemgrep = *maxSemgrep
	}
	if *maxCodeQL > 0 {
		opt.MaxCodeQL = *maxCodeQL
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
	log.Info("coderules scrape done")
}
