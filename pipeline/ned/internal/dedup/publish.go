package dedup

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/pkg/commit"
	"github.com/butbeautifulv/threat_intelligence/pkg/harvest"
	connats "github.com/butbeautifulv/threat_intelligence/pipeline/connector/nats"
	"github.com/butbeautifulv/threat_intelligence/pipeline/ned/internal/transform"
)

// PublishIngest publishes envelopes with JetStream dedup via Nats-Msg-Id (idempotency_key).
func PublishIngest(ctx context.Context, pub *connats.JetStreamPublisher, subject string, envs []*commit.Envelope) error {
	for _, e := range envs {
		if err := pub.PublishJSON(ctx, subject, e); err != nil {
			return err
		}
	}
	return nil
}

// ProcessScrapeMessage runs NED on a scrape envelope and publishes ingest results.
func ProcessScrapeMessage(ctx context.Context, pub *connats.JetStreamPublisher, ingestSubject string, env *harvest.Envelope) error {
	out, err := transform.ScrapeToIngest(ctx, env)
	if err != nil {
		return err
	}
	return PublishIngest(ctx, pub, ingestSubject, out)
}
