package ingest

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/graph/ingest/internal/components"
	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/nats-io/nats.go"
)

func TestValidateEnvelopeSource_mismatch(t *testing.T) {
	env := &commit.Envelope{
		SchemaVersion: commit.CurrentSchemaVersion,
		Source:        commit.SourceVuln,
		Kind:          commit.KindEngageToolRun,
		IdempotencyKey: "x",
		Payload:        json.RawMessage(`{}`),
	}
	if err := validateEnvelopeSource(env); err == nil {
		t.Fatal("expected source/kind mismatch error")
	}
}

func TestValidateEnvelopeSource_engageOK(t *testing.T) {
	payload, _ := json.Marshal(commit.EngageToolRunPayload{
		Tool: "nmap", Target: "127.0.0.1", Subject: "t", Success: true, At: "2026-05-16T00:00:00Z",
	})
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey("nmap", "127.0.0.1", "2026-05-16T00:00:00Z"),
		Payload:        payload,
	}
	if err := validateEnvelopeSource(env); err != nil {
		t.Fatal(err)
	}
}

func TestHandleMsg_routesEngageToApplier(t *testing.T) {
	var called bool
	rt := &components.Runtime{
		Apply: components.DomainAppliers{
			Engage: func(_ context.Context, e *commit.Envelope) error {
				called = true
				if e.Kind != commit.KindEngageToolRun {
					t.Fatalf("kind %s", e.Kind)
				}
				return nil
			},
		},
	}
	payload, _ := json.Marshal(commit.EngageToolRunPayload{
		Tool: "nmap", Target: "127.0.0.1", Subject: "t", Success: true, At: "2026-05-16T00:00:00Z",
	})
	env := commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey("nmap", "127.0.0.1", "2026-05-16T00:00:00Z"),
		Payload:        payload,
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := handleMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, rt); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Engage applier not called")
	}
}
