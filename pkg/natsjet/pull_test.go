package natsjet

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func startTestNATSPull(t *testing.T) (*server.Server, string) {
	t.Helper()
	opts := &server.Options{Port: -1, JetStream: true}
	s, err := server.NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	go s.Start()
	if !s.ReadyForConnections(5 * time.Second) {
		t.Fatal("nats server not ready")
	}
	t.Cleanup(func() { s.Shutdown() })
	return s, s.ClientURL()
}

func TestRunPullLoop_processesAndAcks(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_TEST", []string{"pull.test.>"}); err != nil {
		t.Fatal(err)
	}
	if _, err := js.Publish("pull.test.one", []byte("hi")); err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.test.>", "pull-test-durable", nats.BindStream("PULL_TEST"))
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()

	var got atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
			Batch:   1,
			MaxWait: 500 * time.Millisecond,
		}, func(_ context.Context, m *nats.Msg) error {
			if string(m.Data) != "hi" {
				t.Errorf("data %q", m.Data)
			}
			got.Add(1)
			cancel()
			return nil
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
	if got.Load() != 1 {
		t.Fatalf("got %d messages", got.Load())
	}
}

func TestRunPullLoop_cancelWithoutReturnErr(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_NR", []string{"pull.nr.>"}); err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.nr.>", "pull-nr", nats.BindStream("PULL_NR"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 20 * time.Millisecond,
	}, func(context.Context, *nats.Msg) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestRunPullLoop_fetchTimeoutAndNilLog(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_TO", []string{"pull.to.>"}); err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.to.>", "pull-to", nats.BindStream("PULL_TO"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := RunPullLoop(ctx, nil, sub, PullLoopOpts{Batch: 1, MaxWait: 30 * time.Millisecond}, func(context.Context, *nats.Msg) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
