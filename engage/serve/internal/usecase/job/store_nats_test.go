package job

import (
	"context"
	"testing"
	"time"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
)

func startTestNATS(t *testing.T) string {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	return srv.ClientURL()
}

func TestNATSStore_enqueueConsumeAck(t *testing.T) {
	url := startTestNATS(t)
	store, err := NewNATSStore(url)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	j := &domain.Job{
		ID: "job-nats-1", ToolName: "nmap_scan", Target: "127.0.0.1",
		Status: domain.StatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := store.Put(j); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id, msg, err := store.FetchPending(ctx, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if id != j.ID {
		t.Fatalf("id=%q want %q", id, j.ID)
	}
	claimed, ok := store.TryClaim(id)
	if !ok || claimed.Status != domain.StatusRunning {
		t.Fatalf("claim failed: %+v", claimed)
	}
	if err := store.Ack(msg); err != nil {
		t.Fatal(err)
	}
	claimed.Status = domain.StatusDone
	if err := store.Put(claimed); err != nil {
		t.Fatal(err)
	}
	got, ok := store.Get(j.ID)
	if !ok || got.Status != domain.StatusDone {
		t.Fatalf("got=%+v ok=%v", got, ok)
	}
}
