// Package natsjet provides JetStream helpers shared by layer connectors.
package natsjet

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go"
)

// Conn wraps a NATS connection and JetStream context.
type Conn struct {
	NC *nats.Conn
	JS nats.JetStreamContext
}

// Connect opens NATS and JetStream.
func Connect(url string) (*Conn, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, err
	}
	return &Conn{NC: nc, JS: js}, nil
}

func (c *Conn) Close() { _ = c.NC.Drain() }

// PublishJSON marshals v, publishes with Nats-Msg-Id, and respects ctx cancellation.
func (c *Conn) PublishJSON(ctx context.Context, subject string, v any, msgID string) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := c.JS.Publish(subject, data, nats.MsgId(msgID)); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

// EnsureStream creates or widens a JetStream stream to include all subjects.
func EnsureStream(js nats.JetStreamContext, name string, subjects []string) error {
	info, err := js.StreamInfo(name)
	if err != nil {
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return err
		}
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     name,
			Subjects: subjects,
			Storage:  nats.FileStorage,
		})
		return err
	}
	for _, want := range subjects {
		found := false
		for _, s := range info.Config.Subjects {
			if s == want {
				found = true
				break
			}
		}
		if found {
			continue
		}
		cfg := info.Config
		cfg.Subjects = subjects
		_, err = js.UpdateStream(&cfg)
		return err
	}
	return nil
}
