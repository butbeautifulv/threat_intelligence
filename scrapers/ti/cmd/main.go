package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"
	"ti/internal/feeds"
	"ti/internal/natspub"
)

func main() {
	var (
		input     = flag.String("input", "", "optional path to TI JSONL file")
		feedsFlag = flag.String("feeds", "", "comma-separated public feeds: kev,pt,urlhaus,threatfox,malwarebazaar,feodo,openphish")
		natsURL   = flag.String("nats-url", envOr("NATS_URL", "nats://localhost:4222"), "NATS server URL")
		tiSubject = flag.String("ti-nats-subject", envOr("TI_NATS_SUBJECT", "ingest.ti.events"), "JetStream publish subject for TI")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigQuit
		cancel()
	}()

	if err := run(ctx, logger, input, feedsFlag, natsURL, tiSubject); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, logger *slog.Logger, input, feedsFlag, natsURL, tiSubject *string) error {
	pub, err := ingestpub.ConnectJetStreamAndStream(strings.TrimSpace(*natsURL))
	if err != nil {
		return err
	}
	defer pub.Close()

	repo := natspub.New(pub, strings.TrimSpace(*tiSubject))

	if *feedsFlag != "" {
		kinds := strings.Split(*feedsFlag, ",")
		runner := feeds.NewRunner(repo, logger)
		if err := runner.Run(ctx, kinds); err != nil {
			return err
		}
	}

	if *input != "" {
		f, err := os.Open(*input)
		if err != nil {
			return err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for sc.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			line := bytes.TrimSpace(sc.Bytes())
			if len(line) == 0 {
				continue
			}
			sum := sha256.Sum256(line)
			key := ingestv1.TIJSONLRecordIdempotencyKey(hex.EncodeToString(sum[:]))
			env, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIJSONLRecord, key, ingestv1.TIJSONLRecordPayload{
				Line: json.RawMessage(line),
			})
			if err != nil {
				return err
			}
			if err := pub.PublishJSON(ctx, strings.TrimSpace(*tiSubject), env); err != nil {
				return err
			}
		}
		if err := sc.Err(); err != nil {
			return err
		}
	}

	if *input == "" && *feedsFlag == "" {
		logger.Info("nothing to do: pass --feeds ... and/or --input path/to.jsonl")
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
