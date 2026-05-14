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
	"time"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"
	"ti/internal/feeds"
	"ti/internal/ingest"
	"ti/internal/natspub"
	neo4jstore "ti/internal/storage/neo4j"
	"ti/internal/usecase"
)

func main() {
	var (
		input     = flag.String("input", "", "optional path to TI JSONL file")
		feedsFlag = flag.String("feeds", "", "comma-separated public feeds: kev,pt,urlhaus,threatfox,malwarebazaar,feodo,openphish")

		neo4jURI  = flag.String("neo4j-uri", envOr("NEO4J_URI", "neo4j://localhost:7687"), "neo4j uri")
		neo4jUser = flag.String("neo4j-user", envOr("NEO4J_USER", "neo4j"), "neo4j username")
		neo4jPass = flag.String("neo4j-pass", envOr("NEO4J_PASS", "neo4jpassword"), "neo4j password")
		neo4jDB   = flag.String("neo4j-db", envOr("NEO4J_DB", "neo4j"), "neo4j database")

		ingestMode = flag.String("ingest-mode", strings.ToLower(strings.TrimSpace(envOr("INGEST_MODE", "direct"))), "direct or nats")
		natsURL    = flag.String("nats-url", envOr("NATS_URL", "nats://localhost:4222"), "NATS url for INGEST_MODE=nats")
		tiSubject  = flag.String("ti-nats-subject", envOr("TI_NATS_SUBJECT", "ingest.ti.events"), "JetStream publish subject for TI")
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

	mode := strings.ToLower(strings.TrimSpace(*ingestMode))
	if mode != "nats" {
		mode = "direct"
	}

	if mode == "nats" {
		if err := runNATS(ctx, logger, input, feedsFlag, natsURL, tiSubject); err != nil {
			log.Fatal(err)
		}
		return
	}

	store, err := neo4jstore.New(ctx, neo4jstore.Config{
		URI:      *neo4jURI,
		Username: *neo4jUser,
		Password: *neo4jPass,
		Database: *neo4jDB,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cctx, ccancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer ccancel()
		_ = store.Close(cctx)
	}()

	if *feedsFlag != "" {
		kinds := strings.Split(*feedsFlag, ",")
		runner := feeds.NewRunner(store, logger)
		if err := runner.Run(ctx, kinds); err != nil {
			log.Fatal(err)
		}
	}

	if *input != "" {
		stream, closeFn, err := ingest.NewStreamFromFile(*input)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = closeFn() }()

		uc := usecase.NewIngestor(store, logger)
		if err := uc.IngestJSONL(ctx, stream); err != nil {
			log.Fatal(err)
		}
	}

	if *input == "" && *feedsFlag == "" {
		logger.Info("nothing to do: pass --feeds kev,urlhaus,threatfox,... and/or --input path/to.jsonl")
	}
}

func runNATS(ctx context.Context, logger *slog.Logger, input, feedsFlag, natsURL, tiSubject *string) error {
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
