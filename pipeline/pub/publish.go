package ingestpub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/nats-io/nats.go"
)

// JetStreamPublisher publishes ingestv1 envelopes with deduplication header.
type JetStreamPublisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func ConnectJetStream(url string) (*JetStreamPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, err
	}
	return &JetStreamPublisher{nc: nc, js: js}, nil
}

func (p *JetStreamPublisher) Close() { _ = p.nc.Drain() }

// PublishJSON marshals the envelope and publishes to subject with Nats-Msg-Id.
func (p *JetStreamPublisher) PublishJSON(ctx context.Context, subject string, env *ingestv1.Envelope) error {
	if err := env.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if _, err := p.js.Publish(subject, data, nats.MsgId(env.IdempotencyKey)); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

// EnsureIngestStream creates or updates the INGEST stream to accept all ingest.* subjects.
func EnsureIngestStream(js nats.JetStreamContext) error {
	info, err := js.StreamInfo("INGEST")
	if err != nil {
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return err
		}
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     "INGEST",
			Subjects: []string{"ingest.>"},
			Storage:  nats.FileStorage,
		})
		return err
	}
	// Widen legacy streams that only matched ingest.appsec.>
	hasIngestAll := false
	for _, s := range info.Config.Subjects {
		if s == "ingest.>" {
			hasIngestAll = true
			break
		}
	}
	if hasIngestAll {
		return nil
	}
	cfg := info.Config
	cfg.Subjects = []string{"ingest.>"}
	_, err = js.UpdateStream(&cfg)
	return err
}

// EnsureAppSecStream is kept for callers; it now ensures the unified ingest stream.
func EnsureAppSecStream(js nats.JetStreamContext) error {
	return EnsureIngestStream(js)
}

// ConnectJetStreamAndStream connects and ensures INGEST stream exists.
func ConnectJetStreamAndStream(url string) (*JetStreamPublisher, error) {
	pub, err := ConnectJetStream(url)
	if err != nil {
		return nil, err
	}
	if err := EnsureAppSecStream(pub.js); err != nil {
		pub.Close()
		return nil, fmt.Errorf("ingest stream: %w", err)
	}
	return pub, nil
}

// EnsureBothStreams ensures SCRAPE and INGEST streams (for pipeline-worker).
func EnsureBothStreams(js nats.JetStreamContext) error {
	if err := EnsureScrapeStream(js); err != nil {
		return err
	}
	return EnsureIngestStream(js)
}

// EnsureScrapeStream creates or updates SCRAPE (scrape.>).
func EnsureScrapeStream(js nats.JetStreamContext) error {
	info, err := js.StreamInfo("SCRAPE")
	if err != nil {
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return err
		}
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     "SCRAPE",
			Subjects: []string{"scrape.>"},
			Storage:  nats.FileStorage,
		})
		return err
	}
	for _, s := range info.Config.Subjects {
		if s == "scrape.>" {
			return nil
		}
	}
	cfg := info.Config
	cfg.Subjects = []string{"scrape.>"}
	_, err = js.UpdateStream(&cfg)
	return err
}
