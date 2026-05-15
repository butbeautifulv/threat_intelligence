package scrapepub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"github.com/nats-io/nats.go"
)

// JetStreamPublisher publishes scrapev1 envelopes to scrape.> subjects.
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

func (p *JetStreamPublisher) PublishJSON(ctx context.Context, subject string, env *scrapev1.Envelope) error {
	if err := env.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if _, err := p.js.Publish(subject, data, nats.MsgId(env.ContentKey)); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

// EnsureScrapeStream creates or updates the SCRAPE stream (scrape.>).
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

// ConnectJetStreamAndStream connects and ensures SCRAPE stream exists.
func ConnectJetStreamAndStream(url string) (*JetStreamPublisher, error) {
	pub, err := ConnectJetStream(url)
	if err != nil {
		return nil, err
	}
	if err := EnsureScrapeStream(pub.js); err != nil {
		pub.Close()
		return nil, fmt.Errorf("scrape stream: %w", err)
	}
	return pub, nil
}
