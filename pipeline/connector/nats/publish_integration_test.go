package nats

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	natsgo "github.com/nats-io/nats.go"
)

func startJetStreamTestServer(t *testing.T) natsgo.JetStreamContext {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	nc, err := natsgo.Connect(srv.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	return js
}

func startJetStreamPublisher(t *testing.T) *JetStreamPublisher {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	pub, err := ConnectJetStream(srv.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pub.Close)
	return pub
}

func TestEnsureIngestStream_createsStream(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureIngestStream(js); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo("INGEST")
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Config.Subjects) != 1 || info.Config.Subjects[0] != "ingest.>" {
		t.Fatalf("subjects %#v", info.Config.Subjects)
	}
}

func TestEnsureBothStreams_scrapeAndIngest(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureBothStreams(js); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SCRAPE", "INGEST"} {
		if _, err := js.StreamInfo(name); err != nil {
			t.Fatalf("stream %s: %v", name, err)
		}
	}
}

func TestEnsureScrapeStream_createsStream(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureScrapeStream(js); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo(natsjet.StreamScrape)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Config.Subjects) != 1 || info.Config.Subjects[0] != "scrape.>" {
		t.Fatalf("subjects %#v", info.Config.Subjects)
	}
}

func TestEnsureEngageEventsStream_createsStream(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureEngageEventsStream(js); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo(natsjet.StreamEngageEvents)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Config.Subjects) != 1 || info.Config.Subjects[0] != "engage.events.>" {
		t.Fatalf("subjects %#v", info.Config.Subjects)
	}
}

func TestEnsureAppSecStream_createsIngest(t *testing.T) {
	js := startJetStreamTestServer(t)
	if err := EnsureAppSecStream(js); err != nil {
		t.Fatal(err)
	}
	if _, err := js.StreamInfo(natsjet.StreamIngest); err != nil {
		t.Fatal(err)
	}
}

func TestConnectJetStreamAndStream_ensuresIngest(t *testing.T) {
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	pub, err := ConnectJetStreamAndStream(srv.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pub.Close)
	if _, err := pub.conn.JS.StreamInfo(natsjet.StreamIngest); err != nil {
		t.Fatal(err)
	}
}

func TestPublishJSON_andPublishCommit_success(t *testing.T) {
	pub := startJetStreamPublisher(t)
	if err := EnsureIngestStream(pub.conn.JS); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	key := "engage:run:nmap:host:2020-01-01T00:00:00Z"
	pl := map[string]any{"tool": "nmap", "target": "host"}
	if err := pub.PublishCommit(ctx, "ingest.engage.tool_run", commit.SourceEngage, commit.KindEngageToolRun, key, pl); err != nil {
		t.Fatal(err)
	}
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageFinding,
		IdempotencyKey: commit.EngageFindingIdempotencyKey("nuclei", "https://x", "sqli"),
		Payload:        []byte(`{"tool":"nuclei","target":"https://x","title":"sqli"}`),
	}
	if err := pub.PublishJSON(ctx, "ingest.engage.finding", env); err != nil {
		t.Fatal(err)
	}
	info, err := pub.conn.JS.StreamInfo(natsjet.StreamIngest)
	if err != nil {
		t.Fatal(err)
	}
	if info.State.Msgs != 2 {
		t.Fatalf("stream msgs = %d, want 2", info.State.Msgs)
	}
}
