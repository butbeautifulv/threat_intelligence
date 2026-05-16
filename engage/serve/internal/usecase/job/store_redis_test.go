package job

import (
	"context"
	"testing"
	"time"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
	"github.com/alicebob/miniredis/v2"
)

func TestRedisStore_putClaimDone(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	store, err := NewRedisStore("redis://" + mr.Addr())
	if err != nil {
		t.Fatal(err)
	}

	j := &domain.Job{
		ID: "job-1", ToolName: "nmap_scan", Target: "127.0.0.1",
		Status: domain.StatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := store.Put(j); err != nil {
		t.Fatal(err)
	}
	pending, _ := store.ListPending()
	if len(pending) != 1 {
		t.Fatalf("pending: %d", len(pending))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	id, err := store.BlockingPop(ctx, time.Second)
	if err != nil || id != "job-1" {
		t.Fatalf("pop: id=%q err=%v", id, err)
	}
	got, ok := store.Get(id)
	if !ok {
		t.Fatal("missing job")
	}
	got.Status = domain.StatusRunning
	_ = store.Put(got)
	got.Status = domain.StatusDone
	_ = store.Put(got)
	final, _ := store.Get(id)
	if final.Status != domain.StatusDone {
		t.Fatalf("status %s", final.Status)
	}
}
