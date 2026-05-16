package job

import (
	"context"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
)

func TestQueue_enqueue(t *testing.T) {
	specs := []tool.Spec{
		{Name: "echo_job", Category: "network", Binary: "echo", ArgsTemplate: []string{"{target}"}, TimeoutSec: 5, Enabled: true},
	}
	r := &toolsuc.Runner{Registry: tools.NewRegistry(specs), Exec: &runner.Executor{}}
	q := NewQueue(r, 1)
	j, err := q.Enqueue("echo_job", "ok", "")
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

func TestRunWorker_cancel(t *testing.T) {
	q := NewQueue(&toolsuc.Runner{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := q.RunWorker(ctx); err == nil {
		t.Fatal("expected cancel error")
	}
}
