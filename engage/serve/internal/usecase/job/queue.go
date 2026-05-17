package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/job"
	"github.com/butbeautifulv/veil/engage/serve/internal/telemetry"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

const (
	ModeMemory = "memory"
	ModeFile   = "file"
	ModeRedis  = "redis"
	ModeNats   = "nats"
)

// Queue runs tool jobs asynchronously via memory or file store.
type Queue struct {
	store        Store
	tools        *toolsuc.Runner
	mode         string
	pollInterval time.Duration
	concurrency  int
}

type QueueOption func(*Queue)

func WithStore(s Store) QueueOption {
	return func(q *Queue) { q.store = s }
}

func WithMode(mode string) QueueOption {
	return func(q *Queue) { q.mode = mode }
}

func WithPollInterval(d time.Duration) QueueOption {
	return func(q *Queue) { q.pollInterval = d }
}

func WithConcurrency(n int) QueueOption {
	return func(q *Queue) { q.concurrency = n }
}

func NewQueue(tools *toolsuc.Runner, opts ...QueueOption) *Queue {
	q := &Queue{
		store:        NewMemoryStore(),
		tools:        tools,
		mode:         ModeMemory,
		pollInterval: time.Second,
		concurrency:  2,
	}
	for _, opt := range opts {
		opt(q)
	}
	if q.pollInterval <= 0 {
		q.pollInterval = time.Second
	}
	if q.concurrency <= 0 {
		q.concurrency = 1
	}
	return q
}

func (q *Queue) Enqueue(toolName, target, subject string, parameters map[string]string) (*domain.Job, error) {
	if q.tools == nil {
		return nil, fmt.Errorf("tool runner not configured")
	}
	if _, err := q.tools.Registry.MustGet(toolName); err != nil {
		return nil, err
	}
	id := fmt.Sprintf("job-%d", time.Now().UnixNano())
	j := &domain.Job{
		ID:         id,
		ToolName:   toolName,
		Target:     target,
		Subject:    subject,
		Parameters: parameters,
		Status:     domain.StatusPending,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := q.store.Put(j); err != nil {
		return nil, err
	}
	if q.mode == ModeMemory {
		go q.execute(j.Subject, j.ID)
	}
	got, ok := q.store.Get(id)
	if !ok {
		return j, nil
	}
	return got, nil
}

func (q *Queue) Get(id string) (*domain.Job, bool) {
	return q.store.Get(id)
}

func (q *Queue) List(status domain.Status, limit int) ([]*domain.Job, error) {
	return q.store.ListByStatus(status, limit)
}

func (q *Queue) Cancel(id string) error {
	j, ok := q.store.Get(id)
	if !ok {
		return fmt.Errorf("job not found")
	}
	if j.Status != domain.StatusPending {
		return fmt.Errorf("cannot cancel job in status %s", j.Status)
	}
	j.Status = domain.StatusCancelled
	j.UpdatedAt = time.Now().UTC()
	return q.store.Put(j)
}

func (q *Queue) CountByStatus(status domain.Status) (int, error) {
	return q.store.CountByStatus(status)
}

func (q *Queue) execute(subject, id string) {
	j, ok := q.store.TryClaim(id)
	if !ok {
		return
	}
	q.runTool(subject, j)
	_ = q.store.Put(j)
}

func (q *Queue) runTool(subject string, j *domain.Job) {
	if subject == "" {
		subject = j.Subject
	}
	res := q.tools.Run(context.Background(), subject, j.ToolName, contract.ToolRunRequest{
		Target:     j.Target,
		Parameters: j.Parameters,
	})
	j.Output = res.Output
	j.Error = res.Error
	j.UpdatedAt = time.Now().UTC()
	if res.Success {
		j.Status = domain.StatusDone
		telemetry.RecordJob("done")
	} else {
		j.Status = domain.StatusFailed
		telemetry.RecordJob("failed")
	}
}

// RunWorker polls the store for pending jobs until ctx is cancelled.
func (q *Queue) RunWorker(ctx context.Context) error {
	if q.mode == ModeNats {
		return q.runNATSWorker(ctx)
	}
	if q.mode == ModeRedis {
		return q.runRedisWorker(ctx)
	}
	if q.mode != ModeFile {
		<-ctx.Done()
		return ctx.Err()
	}
	ticker := time.NewTicker(q.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			q.processPending(ctx)
		}
	}
}

func (q *Queue) runRedisWorker(ctx context.Context) error {
	rs, ok := q.store.(*RedisStore)
	if !ok {
		return fmt.Errorf("redis mode requires RedisStore")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		id, err := rs.BlockingPop(ctx, q.pollInterval)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		j, ok := q.store.TryClaim(id)
		if !ok {
			continue
		}
		q.runTool(j.Subject, j)
		_ = q.store.Put(j)
	}
}

func (q *Queue) runNATSWorker(ctx context.Context) error {
	ns, ok := q.store.(*NATSStore)
	if !ok {
		return fmt.Errorf("nats mode requires NATSStore")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		id, msg, err := ns.FetchPending(ctx, q.pollInterval)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		j, ok := q.store.TryClaim(id)
		if !ok {
			_ = ns.Ack(msg)
			continue
		}
		q.runTool(j.Subject, j)
		_ = q.store.Put(j)
		_ = ns.Ack(msg)
	}
}

func (q *Queue) processPending(ctx context.Context) {
	pending, err := q.store.ListPending()
	if err != nil || len(pending) == 0 {
		return
	}
	sem := make(chan struct{}, q.concurrency)
	var wg sync.WaitGroup
	for _, p := range pending {
		select {
		case <-ctx.Done():
			return
		default:
		}
		wg.Add(1)
		go func(jobID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			j, ok := q.store.TryClaim(jobID)
			if !ok {
				return
			}
			q.runTool(j.Subject, j)
			_ = q.store.Put(j)
		}(p.ID)
	}
	wg.Wait()
}
