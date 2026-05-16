package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/dedup"
	"github.com/nats-io/nats.go"
)

// RunPullLoop consumes scrape JetStream messages until ctx is canceled.
func RunPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, pub dedup.IngestPublisher, cfg config.Config) error {
	return natsjet.RunPullLoop(ctx, log, sub, natsjet.PullLoopOpts{
		Batch:    batch,
		MaxWait:  maxWait,
		NakDelay: 2 * time.Second,
		StopLog:  "pipeline consumer stopped",
	}, func(ctx context.Context, m *nats.Msg) error {
		return handleScrapeMsg(ctx, log, m, pub, cfg)
	})
}

func handleScrapeMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, pub dedup.IngestPublisher, cfg config.Config) error {
	var env harvest.Envelope
	if err := json.Unmarshal(m.Data, &env); err != nil {
		return fmt.Errorf("decode harvest: %w", err)
	}
	if err := env.Validate(); err != nil {
		return err
	}
	subj := cfg.IngestSubjectFor(env.Source)
	if err := dedup.ProcessScrapeMessage(ctx, pub, subj, &env); err != nil {
		return err
	}
	log.Debug("pipeline processed", slog.String("source", env.Source), slog.String("kind", env.Kind))
	return nil
}
