package ingestpub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
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

// EnsureAppSecStream creates the INGEST stream if missing (idempotent).
func EnsureAppSecStream(js nats.JetStreamContext) error {
	if _, err := js.StreamInfo("INGEST"); err != nil {
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return err
		}
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     "INGEST",
			Subjects: []string{"ingest.appsec.>"},
			Storage:  nats.FileStorage,
		})
		return err
	}
	return nil
}

// ConnectJetStreamAndStream connects and ensures stream exists.
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
