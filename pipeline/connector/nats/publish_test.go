package nats

import (
	"context"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
)

func TestConnectJetStream_invalidURL(t *testing.T) {
	t.Parallel()
	_, err := ConnectJetStream("not-a-valid-nats-url://")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectJetStreamAndStream_connectError(t *testing.T) {
	t.Parallel()
	_, err := ConnectJetStreamAndStream("not-a-valid-nats-url://")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectJetStreamAndStream_ensureStreamError(t *testing.T) {
	opts := &server.Options{JetStream: false, Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	_, err := ConnectJetStreamAndStream(srv.ClientURL())
	if err == nil || !strings.Contains(err.Error(), "ingest stream:") {
		t.Fatalf("err = %v, want ingest stream wrap", err)
	}
}

func TestPublishJSON_errors(t *testing.T) {
	t.Parallel()
	pub := &JetStreamPublisher{conn: &natsjet.Conn{}}
	ctx := context.Background()

	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: "",
		Payload:        []byte(`{}`),
	}
	err := pub.PublishJSON(ctx, "ingest.engage.tool_run", env)
	if err == nil || !strings.Contains(err.Error(), "idempotency_key") {
		t.Fatalf("err = %v, want idempotency_key validation", err)
	}
}

func TestPublishCommit_errors(t *testing.T) {
	t.Parallel()
	pub := &JetStreamPublisher{conn: &natsjet.Conn{}}
	ctx := context.Background()

	err := pub.PublishCommit(ctx, "ingest.engage.tool_run", commit.SourceEngage, commit.KindEngageToolRun, "", map[string]any{"tool": "nmap"})
	if err == nil || !strings.Contains(err.Error(), "idempotency_key") {
		t.Fatalf("err = %v, want idempotency_key validation", err)
	}

	err = pub.PublishCommit(ctx, "ingest.engage.tool_run", commit.SourceEngage, commit.KindEngageToolRun, "key", make(chan int))
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("err = %v, want marshal error", err)
	}
}
