package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/butbeautifulv/threat_intelligence/pkg/harvest"
	connats "github.com/butbeautifulv/threat_intelligence/pipeline/connector/nats"
	"github.com/butbeautifulv/threat_intelligence/pipeline/ned/internal/config"
	"github.com/butbeautifulv/threat_intelligence/pipeline/ned/internal/dedup"
	"github.com/nats-io/nats.go"
)

// RunPullLoop consumes scrape JetStream messages until ctx is canceled.
func RunPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, pub *connats.JetStreamPublisher, cfg config.Config) error {
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
			if err := handleScrapeMsg(ctx, log, m, pub, cfg); err != nil {
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

func handleScrapeMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, pub *connats.JetStreamPublisher, cfg config.Config) error {
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
