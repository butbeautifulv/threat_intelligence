package components

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	connats "github.com/butbeautifulv/veil/pipeline/connector/nats"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	natsgo "github.com/nats-io/nats.go"
)

func startTestNATS(t *testing.T) string {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	return srv.ClientURL()
}

func testConfig(url string) config.Config {
	return config.Config{
		NATSURL:       url,
		ScrapeStream:  "SCRAPE",
		ScrapeDurable: "pipeline-worker-test",
		ScrapeSubject: "scrape.>",
		IngestPublish: "ingest.events",
		Batch:         1,
		MaxWait:       50 * time.Millisecond,
	}
}

func TestInit_successAndShutdown(t *testing.T) {
	url := startTestNATS(t)
	ctx := context.Background()
	rt, err := Init(ctx, testConfig(url), slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	rt.Shutdown()
}

func TestInit_invalidNATS(t *testing.T) {
	_, err := Init(context.Background(), testConfig("nats://127.0.0.1:1"), slog.Default())
	if err == nil {
		t.Fatal("expected connect error")
	}
}

func TestRuntime_Run_cancel(t *testing.T) {
	url := startTestNATS(t)
	rt, err := Init(context.Background(), testConfig(url), slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Shutdown()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := rt.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestShutdown_nilSafe(t *testing.T) {
	var rt Runtime
	rt.Shutdown()
}

func TestInit_jetStreamError(t *testing.T) {
	url := startTestNATS(t)
	old := jetStreamContext
	jetStreamContext = func(*natsgo.Conn) (natsgo.JetStreamContext, error) {
		return nil, errors.New("js unavailable")
	}
	t.Cleanup(func() { jetStreamContext = old })
	_, err := Init(context.Background(), testConfig(url), slog.Default())
	if err == nil || !strings.Contains(err.Error(), "jetstream") {
		t.Fatalf("got %v", err)
	}
}

func TestInit_ensureStreamsError(t *testing.T) {
	url := startTestNATS(t)
	old := ensureBothStreams
	ensureBothStreams = func(natsgo.JetStreamContext) error { return errors.New("streams fail") }
	t.Cleanup(func() { ensureBothStreams = old })
	_, err := Init(context.Background(), testConfig(url), slog.Default())
	if err == nil || !strings.Contains(err.Error(), "streams") {
		t.Fatalf("got %v", err)
	}
}

func TestInit_connectPublisherError(t *testing.T) {
	url := startTestNATS(t)
	old := connectIngestPublisher
	connectIngestPublisher = func(string) (*connats.JetStreamPublisher, error) {
		return nil, errors.New("publisher fail")
	}
	t.Cleanup(func() { connectIngestPublisher = old })
	_, err := Init(context.Background(), testConfig(url), slog.Default())
	if err == nil {
		t.Fatal("expected publisher error")
	}
}

func TestInit_pullSubscribeBadStream(t *testing.T) {
	url := startTestNATS(t)
	cfg := testConfig(url)
	cfg.ScrapeStream = "NOEXIST"
	_, err := Init(context.Background(), cfg, slog.Default())
	if err == nil {
		t.Fatal("expected pull subscribe error")
	}
}

func TestRun_initFails(t *testing.T) {
	t.Setenv("NATS_URL", "nats://127.0.0.1:1")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Run(ctx, slog.Default()); err == nil {
		t.Fatal("expected init error")
	}
}

func TestRun_topLevelCanceled(t *testing.T) {
	url := startTestNATS(t)
	t.Setenv("NATS_URL", url)
	t.Setenv("NATS_SCRAPE_DURABLE", "pipeline-run-top-cancel")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := Run(ctx, slog.Default())
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("run: %v", err)
	}
}

