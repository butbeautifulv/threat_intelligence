package nats

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	natsgo "github.com/nats-io/nats.go"
)

func startIngestPublisher(t *testing.T) *JetStreamPublisher {
	t.Helper()
	pub := startJetStreamPublisher(t)
	if err := EnsureIngestStream(pub.conn.JS); err != nil {
		t.Fatal(err)
	}
	return pub
}

func fetchOneCommitEnvelope(t *testing.T, pub *JetStreamPublisher, subject string) commit.Envelope {
	t.Helper()
	sub, err := pub.conn.JS.SubscribeSync(subject, natsgo.BindStream(natsjet.StreamIngest))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	msg, err := sub.NextMsg(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	var env commit.Envelope
	if err := json.Unmarshal(msg.Data, &env); err != nil {
		t.Fatal(err)
	}
	return env
}

func TestHandleEngageFinding_publishesIngestEnvelope(t *testing.T) {
	pub := startIngestPublisher(t)
	wire := engageevents.FindingEvent{
		Tool: "nuclei", Target: "https://example.com", Title: "xss", Severity: "high",
		Description: "reflected xss",
	}
	b, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	m := &natsgo.Msg{Subject: "engage.events.finding", Data: b}
	subject := "ingest.engage.finding"

	if err := handleEngageFinding(context.Background(), pub, subject, m); err != nil {
		t.Fatal(err)
	}

	env := fetchOneCommitEnvelope(t, pub, subject)
	if env.Kind != commit.KindEngageFinding || env.Source != commit.SourceEngage {
		t.Fatalf("env kind/source = %q/%q", env.Kind, env.Source)
	}
	if env.IdempotencyKey != commit.EngageFindingIdempotencyKey(wire.Tool, wire.Target, wire.Title) {
		t.Fatalf("idempotency key = %q", env.IdempotencyKey)
	}
	var pl commit.EngageFindingPayload
	if err := json.Unmarshal(env.Payload, &pl); err != nil {
		t.Fatal(err)
	}
	if pl.Tool != wire.Tool || pl.Title != wire.Title || pl.Severity != wire.Severity {
		t.Fatalf("payload = %+v", pl)
	}
}

func TestHandleEngageFinding_errors(t *testing.T) {
	pub := startIngestPublisher(t)
	ctx := context.Background()
	subject := "ingest.engage.finding"

	if err := handleEngageFinding(ctx, pub, subject, &natsgo.Msg{Data: []byte("{")}); err == nil {
		t.Fatal("expected unmarshal error")
	}

	wire := engageevents.FindingEvent{Tool: "  ", Target: "t", Title: "x"}
	b, _ := json.Marshal(wire)
	if err := handleEngageFinding(ctx, pub, subject, &natsgo.Msg{Data: b}); err == nil || !strings.Contains(err.Error(), "empty tool or title") {
		t.Fatalf("empty tool err = %v", err)
	}

	wire = engageevents.FindingEvent{Tool: "nuclei", Target: "t", Title: "  "}
	b, _ = json.Marshal(wire)
	if err := handleEngageFinding(ctx, pub, subject, &natsgo.Msg{Data: b}); err == nil || !strings.Contains(err.Error(), "empty tool or title") {
		t.Fatalf("empty title err = %v", err)
	}
}

func TestHandleEngageFinding_marshalError(t *testing.T) {
	old := engageJSONMarshal
	t.Cleanup(func() { engageJSONMarshal = old })
	engageJSONMarshal = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }

	pub := startIngestPublisher(t)
	wire := engageevents.FindingEvent{Tool: "nuclei", Target: "https://x", Title: "sqli"}
	b, _ := json.Marshal(wire)
	err := handleEngageFinding(context.Background(), pub, "ingest.engage.finding", &natsgo.Msg{Data: b})
	if err == nil || err.Error() != "marshal failed" {
		t.Fatalf("err = %v, want marshal failed", err)
	}
}

func TestHandleEngageFinding_publishError(t *testing.T) {
	pub := startIngestPublisher(t)
	pub.Close()
	wire := engageevents.FindingEvent{Tool: "nuclei", Target: "https://x", Title: "sqli"}
	b, _ := json.Marshal(wire)
	err := handleEngageFinding(context.Background(), pub, "ingest.engage.finding", &natsgo.Msg{Data: b})
	if err == nil {
		t.Fatal("expected publish error on closed publisher")
	}
}

func TestHandleEngageToolRun_publishesIngestEnvelope(t *testing.T) {
	pub := startIngestPublisher(t)
	at := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	wire := engageevents.AuditEvent{
		Tool: "nmap", Target: "127.0.0.1", Subject: "port-scan", Success: true, At: at,
	}
	b, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	m := &natsgo.Msg{Subject: "engage.events.audit", Data: b}
	subject := "ingest.engage.tool_run"

	if err := handleEngageToolRun(context.Background(), pub, subject, m); err != nil {
		t.Fatal(err)
	}

	env := fetchOneCommitEnvelope(t, pub, subject)
	atStr := at.UTC().Format(time.RFC3339)
	if env.Kind != commit.KindEngageToolRun || env.Source != commit.SourceEngage {
		t.Fatalf("env kind/source = %q/%q", env.Kind, env.Source)
	}
	if env.IdempotencyKey != commit.EngageToolRunIdempotencyKey(wire.Tool, wire.Target, atStr) {
		t.Fatalf("idempotency key = %q", env.IdempotencyKey)
	}
	var pl commit.EngageToolRunPayload
	if err := json.Unmarshal(env.Payload, &pl); err != nil {
		t.Fatal(err)
	}
	if pl.Tool != wire.Tool || pl.Target != wire.Target || pl.At != atStr || !pl.Success {
		t.Fatalf("payload = %+v", pl)
	}
}

func TestHandleEngageToolRun_zeroAt(t *testing.T) {
	pub := startIngestPublisher(t)
	wire := engageevents.AuditEvent{
		Tool: "nmap", Target: "127.0.0.1", Subject: "scan", Success: false,
	}
	b, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	subject := "ingest.engage.tool_run"
	if err := handleEngageToolRun(context.Background(), pub, subject, &natsgo.Msg{Data: b}); err != nil {
		t.Fatal(err)
	}
	env := fetchOneCommitEnvelope(t, pub, subject)
	var pl commit.EngageToolRunPayload
	if err := json.Unmarshal(env.Payload, &pl); err != nil {
		t.Fatal(err)
	}
	if pl.At == "" {
		t.Fatal("expected non-empty At when wire At is zero")
	}
}

func TestHandleEngageToolRun_errors(t *testing.T) {
	pub := startIngestPublisher(t)
	ctx := context.Background()
	subject := "ingest.engage.tool_run"

	if err := handleEngageToolRun(ctx, pub, subject, &natsgo.Msg{Data: []byte("not-json")}); err == nil {
		t.Fatal("expected unmarshal error")
	}

	wire := engageevents.AuditEvent{Tool: "  ", Target: "127.0.0.1"}
	b, _ := json.Marshal(wire)
	if err := handleEngageToolRun(ctx, pub, subject, &natsgo.Msg{Data: b}); err == nil || !strings.Contains(err.Error(), "empty tool") {
		t.Fatalf("empty tool err = %v", err)
	}
}

func TestHandleEngageToolRun_marshalError(t *testing.T) {
	old := engageJSONMarshal
	t.Cleanup(func() { engageJSONMarshal = old })
	engageJSONMarshal = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }

	pub := startIngestPublisher(t)
	wire := engageevents.AuditEvent{Tool: "nmap", Target: "127.0.0.1", At: time.Now().UTC()}
	b, _ := json.Marshal(wire)
	err := handleEngageToolRun(context.Background(), pub, "ingest.engage.tool_run", &natsgo.Msg{Data: b})
	if err == nil || err.Error() != "marshal failed" {
		t.Fatalf("err = %v, want marshal failed", err)
	}
}

func TestHandleEngageToolRun_publishError(t *testing.T) {
	pub := startIngestPublisher(t)
	pub.Close()
	wire := engageevents.AuditEvent{Tool: "nmap", Target: "127.0.0.1", At: time.Now().UTC()}
	b, _ := json.Marshal(wire)
	err := handleEngageToolRun(context.Background(), pub, "ingest.engage.tool_run", &natsgo.Msg{Data: b})
	if err == nil {
		t.Fatal("expected publish error on closed publisher")
	}
}

func TestHandleEngageMsg_routesBySubject(t *testing.T) {
	pub := startIngestPublisher(t)
	ctx := context.Background()

	finding := engageevents.FindingEvent{Tool: "nuclei", Target: "https://x", Title: "sqli"}
	fb, _ := json.Marshal(finding)
	if err := handleEngageMsg(ctx, pub, "ingest.engage.tool_run", "ingest.engage.finding", &natsgo.Msg{
		Subject: "engage.events.finding", Data: fb,
	}); err != nil {
		t.Fatal(err)
	}
	fenv := fetchOneCommitEnvelope(t, pub, "ingest.engage.finding")
	if fenv.Kind != commit.KindEngageFinding {
		t.Fatalf("finding kind = %q", fenv.Kind)
	}

	audit := engageevents.AuditEvent{Tool: "nmap", Target: "host", At: time.Now().UTC()}
	ab, _ := json.Marshal(audit)
	if err := handleEngageMsg(ctx, pub, "ingest.engage.tool_run", "ingest.engage.finding", &natsgo.Msg{
		Subject: "engage.events.audit", Data: ab,
	}); err != nil {
		t.Fatal(err)
	}
	tenv := fetchOneCommitEnvelope(t, pub, "ingest.engage.tool_run")
	if tenv.Kind != commit.KindEngageToolRun {
		t.Fatalf("tool run kind = %q", tenv.Kind)
	}
}

func TestRunEngageEventsConsumer_invalidURL(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := RunEngageEventsConsumer(ctx, nil, "not-a-valid-nats-url://", "", "", "")
	if err == nil {
		t.Fatal("expected connect error")
	}
}

func TestRunEngageEventsConsumer_engageStreamError(t *testing.T) {
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	url := srv.ClientURL()

	setupNC, err := natsgo.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(setupNC.Close)
	setupJS, err := setupNC.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setupJS.AddStream(&natsgo.StreamConfig{Name: "OTHER", Subjects: []string{"engage.events.>"}}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = RunEngageEventsConsumer(ctx, slog.Default(), url, "", "", "")
	if err == nil || !strings.Contains(err.Error(), "engage events stream:") {
		t.Fatalf("err = %v, want engage events stream wrap", err)
	}
}

func TestRunEngageEventsConsumer_pullSubscribeError(t *testing.T) {
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	url := srv.ClientURL()

	setupNC, err := natsgo.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(setupNC.Close)
	setupJS, err := setupNC.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureEngageEventsStream(setupJS); err != nil {
		t.Fatal(err)
	}
	if _, err := setupJS.Subscribe("engage.events.>", func(*natsgo.Msg) {}, natsgo.Durable("engage-events-bridge"), natsgo.BindStream(natsjet.StreamEngageEvents)); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = RunEngageEventsConsumer(ctx, slog.Default(), url, "", "", "")
	if err == nil || !strings.Contains(err.Error(), "pull subscribe:") {
		t.Fatalf("err = %v, want pull subscribe wrap", err)
	}
}

func TestRunEngageEventsConsumer_bridgesFinding(t *testing.T) {
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	url := srv.ClientURL()

	setupNC, err := natsgo.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(setupNC.Close)
	setupJS, err := setupNC.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureEngageEventsStream(setupJS); err != nil {
		t.Fatal(err)
	}
	if err := EnsureIngestStream(setupJS); err != nil {
		t.Fatal(err)
	}
	wire := engageevents.FindingEvent{Tool: "nuclei", Target: "https://bridge.test", Title: "open-port", Severity: "low"}
	b, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setupJS.Publish("engage.events.finding", b); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunEngageEventsConsumer(ctx, slog.Default(), url, "", "", "")
	}()

	ingestSub, err := setupJS.SubscribeSync("ingest.engage.finding", natsgo.BindStream(natsjet.StreamIngest))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ingestSub.Unsubscribe() })
	msg, err := ingestSub.NextMsg(5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	var env commit.Envelope
	if err := json.Unmarshal(msg.Data, &env); err != nil {
		t.Fatal(err)
	}
	if env.Kind != commit.KindEngageFinding {
		t.Fatalf("kind = %q", env.Kind)
	}

	cancel()
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatalf("consumer err = %v, want context.Canceled", err)
	}
}

func TestRunEngageEventsConsumer_handleError(t *testing.T) {
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	url := srv.ClientURL()

	setupNC, err := natsgo.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(setupNC.Close)
	setupJS, err := setupNC.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureEngageEventsStream(setupJS); err != nil {
		t.Fatal(err)
	}
	if _, err := setupJS.Publish("engage.events.audit", []byte("not-json")); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = RunEngageEventsConsumer(ctx, slog.New(slog.DiscardHandler), url, "engage.events.>", "", "")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}
