package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func startPipelinePullNATS(t *testing.T) (string, nats.JetStreamContext) {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	url := srv.ClientURL()
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := natsjet.EnsureScrapeAndIngest(js); err != nil {
		t.Fatal(err)
	}
	return url, js
}

func TestRunPullLoop_processesScrapeMessage(t *testing.T) {
	url, js := startPipelinePullNATS(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	sub, err := js.PullSubscribe("scrape.>", "pipeline-pull-test", nats.BindStream(natsjet.StreamScrape))
	if err != nil {
		t.Fatal(err)
	}
	env := harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceTI,
		Kind:          harvest.KindTIIoCRaw,
		ContentKey:    "ti:ioc:url:https://pull.example/x",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       json.RawMessage(`{"type":"url","value":"https://pull.example/x","source":"t"}`),
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := js.Publish("scrape.ti.ioc", b); err != nil {
		t.Fatal(err)
	}
	rec := &recordIngestPublisher{}
	cfg := config.Config{IngestPublish: "ingest.events"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = RunPullLoop(ctx, slog.Default(), sub, 1, 200*time.Millisecond, rec, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.envs) == 0 {
		t.Fatal("expected ingest publish")
	}
	if rec.envs[0].Kind != commit.KindTIIoC {
		t.Fatalf("kind %s", rec.envs[0].Kind)
	}
}
