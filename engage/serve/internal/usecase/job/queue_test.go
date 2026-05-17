package job

import (
	"context"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
)

func TestQueue_enqueue_memory(t *testing.T) {
	specs := []tool.Spec{
		{Name: "echo_job", Category: "network", Binary: "echo", ArgsTemplate: []string{"{target}"}, TimeoutSec: 5, Enabled: true},
	}
	r := &toolsuc.Runner{Registry: tools.NewRegistry(specs), Exec: &runner.Executor{}}
	q := NewQueue(r, WithMode(ModeMemory))
	j, err := q.Enqueue("echo_job", "ok", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := q.Get(j.ID)
		if ok && got.Status == "done" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("job did not complete")
}

func TestQueue_fileMode_workerRuns(t *testing.T) {
	specs := []tool.Spec{
		{Name: "echo_job", Category: "network", Binary: "echo", ArgsTemplate: []string{"{target}"}, TimeoutSec: 5, Enabled: true},
	}
	r := &toolsuc.Runner{Registry: tools.NewRegistry(specs), Exec: &runner.Executor{}}
	dir := t.TempDir()
	q := NewQueue(r,
		WithStore(NewFileStore(dir)),
		WithMode(ModeFile),
		WithPollInterval(50*time.Millisecond),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = q.RunWorker(ctx)
	}()

	j, err := q.Enqueue("echo_job", "hello", "test-subject", nil)
	if err != nil {
		t.Fatal(err)
	}
	if j.Status != "pending" {
		t.Fatalf("expected pending after enqueue, got %s", j.Status)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := q.Get(j.ID)
		if ok && got.Status == "done" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	got, _ := q.Get(j.ID)
	t.Fatalf("job did not complete in file mode, status=%s", got.Status)
}

func TestRunWorker_cancel_memoryMode(t *testing.T) {
	q := NewQueue(&toolsuc.Runner{}, WithMode(ModeMemory))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := q.RunWorker(ctx); err == nil {
		t.Fatal("expected cancel error")
	}
}
