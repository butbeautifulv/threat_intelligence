package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"ingestpub"

	"github.com/butbeautifulv/threat_intelligence/pipeline/pipeline_worker/internal/handle"
)

func main() {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(rootCtx, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("shutdown complete")
}

func run(rootCtx context.Context, log *slog.Logger) error {
	natsURL := getenv("NATS_URL", "nats://localhost:4222")
	scrapeStream := getenv("NATS_SCRAPE_STREAM", "SCRAPE")
	scrapeDurable := getenv("NATS_SCRAPE_DURABLE", "pipeline-worker")
	scrapeSubject := getenv("NATS_SCRAPE_SUBSCRIBE_SUBJECT", "scrape.>")
	ingestPublish := getenv("NATS_INGEST_PUBLISH_SUBJECT", "ingest.events")
	batch := getenvInt("PIPELINE_BATCH", 10)
	maxWait := getenvDuration("PIPELINE_MAX_WAIT", 5*time.Second)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("jetstream: %w", err)
	}
	if err := ingestpub.EnsureBothStreams(js); err != nil {
		return fmt.Errorf("streams: %w", err)
	}

	ingestPub, err := ingestpub.ConnectJetStream(natsURL)
	if err != nil {
		return err
	}
	defer ingestPub.Close()

	sub, err := js.PullSubscribe(scrapeSubject, scrapeDurable, nats.BindStream(scrapeStream))
	if err != nil {
		return fmt.Errorf("pull subscribe scrape stream=%s: %w", scrapeStream, err)
	}
	defer sub.Unsubscribe()

	log.Info("pipeline-worker started",
		slog.String("nats", natsURL),
		slog.String("scrape_stream", scrapeStream),
		slog.String("scrape_subject", scrapeSubject),
		slog.String("ingest_stream", getenv("NATS_INGEST_STREAM", "INGEST")),
		slog.String("ingest_publish", ingestPublish),
	)

	eg, ctx := errgroup.WithContext(rootCtx)
	eg.Go(func() error {
		return runPullLoop(ctx, log, sub, batch, maxWait, ingestPub, ingestPublish)
	})
	return eg.Wait()
}

func runPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, ingestPub *ingestpub.JetStreamPublisher, ingestSubject string) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("pipeline consumer stopped")
			return nil
		default:
		}
		msgs, err := sub.Fetch(batch, nats.MaxWait(maxWait))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			log.Warn("fetch", slog.String("err", err.Error()))
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second):
			}
			continue
		}
		for _, m := range msgs {
			if err := handleScrapeMsg(ctx, log, m, ingestPub, ingestSubject); err != nil {
				log.Warn("scrape message", slog.String("err", err.Error()))
				_ = m.NakWithDelay(2 * time.Second)
				continue
			}
			if err := m.Ack(); err != nil {
				log.Warn("ack", slog.String("err", err.Error()))
			}
		}
	}
}

func handleScrapeMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, ingestPub *ingestpub.JetStreamPublisher, ingestSubject string) error {
	var env scrapev1.Envelope
	if err := json.Unmarshal(m.Data, &env); err != nil {
		return fmt.Errorf("decode scrapev1: %w", err)
	}
	if err := env.Validate(); err != nil {
		return err
	}
	// Route per-domain ingest subjects when set (e.g. ingest.ds.events).
	subj := ingestSubject
	if s := domainIngestSubject(env.Source); s != "" {
		subj = s
	}
	if err := handle.ProcessMessage(ctx, ingestPub, subj, &env); err != nil {
		return err
	}
	log.Debug("pipeline processed", slog.String("source", env.Source), slog.String("kind", env.Kind))
	return nil
}

func domainIngestSubject(source string) string {
	switch source {
	case scrapev1.SourceDS:
		return strings.TrimSpace(os.Getenv("DS_INGEST_SUBJECT"))
	case scrapev1.SourceTI:
		return strings.TrimSpace(os.Getenv("TI_INGEST_SUBJECT"))
	case scrapev1.SourceVuln:
		return strings.TrimSpace(os.Getenv("VULN_INGEST_SUBJECT"))
	case scrapev1.SourceLola:
		return strings.TrimSpace(os.Getenv("LOLA_INGEST_SUBJECT"))
	case scrapev1.SourceSBOM:
		return strings.TrimSpace(os.Getenv("SBOM_INGEST_SUBJECT"))
	case scrapev1.SourceCoderules:
		return strings.TrimSpace(os.Getenv("CODERULES_INGEST_SUBJECT"))
	case scrapev1.SourceNuclei:
		return strings.TrimSpace(os.Getenv("NUCLEI_INGEST_SUBJECT"))
	default:
		return ""
	}
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvDuration(k string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
