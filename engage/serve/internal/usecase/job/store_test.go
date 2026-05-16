package job

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
)

func TestMemoryStore_putGetClaim(t *testing.T) {
	s := NewMemoryStore()
	j := &domain.Job{
		ID: "job-1", ToolName: "nmap_scan", Target: "127.0.0.1",
		Status: domain.StatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := s.Put(j); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Get("job-1")
	if !ok || got.Status != domain.StatusPending {
		t.Fatalf("get: %+v ok=%v", got, ok)
	}
	claimed, ok := s.TryClaim("job-1")
	if !ok || claimed.Status != domain.StatusRunning {
		t.Fatalf("claim: %+v", claimed)
	}
	stored, _ := s.Get("job-1")
	if stored.Status != domain.StatusRunning {
		t.Fatalf("after claim status=%s", stored.Status)
	}
}

func TestFileStore_putGetClaim(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	j := &domain.Job{
		ID: "job-abc", ToolName: "echo", Target: "x", Subject: "sub",
		Status: domain.StatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := s.Put(j); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "job-abc.json")); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Get("job-abc")
	if !ok || got.Subject != "sub" {
		t.Fatalf("get: %+v", got)
	}
	pending, err := s.ListPending()
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending: %v err=%v", pending, err)
	}
	claimed, ok := s.TryClaim("job-abc")
	if !ok || claimed.Status != domain.StatusRunning {
		t.Fatalf("claim failed")
	}
	pending2, _ := s.ListPending()
	if len(pending2) != 0 {
		t.Fatalf("expected no pending after claim, got %d", len(pending2))
	}
}

func TestMemoryStore_cancel(t *testing.T) {
	s := NewMemoryStore()
	j := &domain.Job{
		ID: "job-cancel", ToolName: "nmap_scan", Target: "x",
		Status: domain.StatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	_ = s.Put(j)
	j.Status = domain.StatusCancelled
	j.UpdatedAt = time.Now().UTC()
	_ = s.Put(j)
	claimed, ok := s.TryClaim("job-cancel")
	if ok {
		t.Fatalf("should not claim cancelled job: %+v", claimed)
	}
}

func TestMemoryStore_listByStatus(t *testing.T) {
	s := NewMemoryStore()
	_ = s.Put(&domain.Job{ID: "a", Status: domain.StatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	_ = s.Put(&domain.Job{ID: "b", Status: domain.StatusDone, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	all, err := s.ListByStatus("", 0)
	if err != nil || len(all) != 2 {
		t.Fatalf("all: %v err=%v", all, err)
	}
}
