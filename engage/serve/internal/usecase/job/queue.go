package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// Queue runs tool jobs asynchronously (in-process; replace with Redis/NATS later).
type Queue struct {
	mu      sync.Mutex
	jobs    map[string]*domain.Job
	tools   *toolsuc.Runner
	workers int
}

func NewQueue(tools *toolsuc.Runner, workers int) *Queue {
	if workers < 1 {
		workers = 1
	}
	return &Queue{
		jobs:    make(map[string]*domain.Job),
		tools:   tools,
		workers: workers,
	}
}

func (q *Queue) Enqueue(toolName, target, subject string) (*domain.Job, error) {
	if q.tools == nil {
		return nil, fmt.Errorf("tool runner not configured")
	}
	if _, err := q.tools.Registry.MustGet(toolName); err != nil {
		return nil, err
	}
	id := fmt.Sprintf("job-%d", time.Now().UnixNano())
	j := &domain.Job{
		ID:        id,
		ToolName:  toolName,
		Target:    target,
		Status:    domain.StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	q.mu.Lock()
	q.jobs[id] = j
	q.mu.Unlock()

	go q.run(subject, j)
	return j, nil
}

func (q *Queue) Get(id string) (*domain.Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	j, ok := q.jobs[id]
	return j, ok
}

func (q *Queue) run(subject string, j *domain.Job) {
	q.setStatus(j, domain.StatusRunning)
	res := q.tools.Run(context.Background(), subject, j.ToolName, contract.ToolRunRequest{Target: j.Target})
	q.mu.Lock()
	defer q.mu.Unlock()
	j.Output = res.Output
	j.Error = res.Error
	j.UpdatedAt = time.Now().UTC()
	if res.Success {
		j.Status = domain.StatusDone
	} else {
		j.Status = domain.StatusFailed
	}
}

func (q *Queue) setStatus(j *domain.Job, st domain.Status) {
	q.mu.Lock()
	defer q.mu.Unlock()
	j.Status = st
	j.UpdatedAt = time.Now().UTC()
}

// RunWorker blocks until ctx is cancelled (drains no external broker; jobs start on Enqueue).
func (q *Queue) RunWorker(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
