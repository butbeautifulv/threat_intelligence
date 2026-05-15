// Package natsensure configures JetStream streams for the graph layer consumer.
package natsensure

import (
	"errors"

	"github.com/nats-io/nats.go"
)

// EnsureIngestStream creates or updates the INGEST stream (ingest.>).
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
	for _, s := range info.Config.Subjects {
		if s == "ingest.>" {
			return nil
		}
	}
	cfg := info.Config
	cfg.Subjects = []string{"ingest.>"}
	_, err = js.UpdateStream(&cfg)
	return err
}
