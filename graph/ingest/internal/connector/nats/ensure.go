// Package nats configures JetStream streams for the ingest worker.
package nats

import (
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/nats-io/nats.go"
)

// EnsureIngestStream creates or updates the INGEST stream (ingest.>).
func EnsureIngestStream(js nats.JetStreamContext) error {
	return natsjet.EnsureStream(js, "INGEST", []string{"ingest.>"})
}
