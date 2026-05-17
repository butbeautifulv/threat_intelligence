package job

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/job"
)

// Store persists async tool jobs (memory or filesystem).
type Store interface {
	Put(j *domain.Job) error
	Get(id string) (*domain.Job, bool)
	ListPending() ([]*domain.Job, error)
	ListByStatus(status domain.Status, limit int) ([]*domain.Job, error)
	CountByStatus(status domain.Status) (int, error)
	TryClaim(id string) (*domain.Job, bool)
}

// MemoryStore is an in-process job store.
type MemoryStore struct {
	mu   sync.Mutex
	jobs map[string]*domain.Job
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{jobs: make(map[string]*domain.Job)}
}

func (s *MemoryStore) Put(j *domain.Job) error {
	if j == nil || j.ID == "" {
		return fmt.Errorf("invalid job")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = cloneJob(j)
	return nil
}

func (s *MemoryStore) Get(id string) (*domain.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	return cloneJob(j), true
}

func (s *MemoryStore) ListPending() ([]*domain.Job, error) {
	return s.ListByStatus(domain.StatusPending, 0)
}

func (s *MemoryStore) ListByStatus(status domain.Status, limit int) ([]*domain.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*domain.Job
	for _, j := range s.jobs {
		if status != "" && j.Status != status {
			continue
		}
		out = append(out, cloneJob(j))
	}
	sortJobsByCreated(out)
	return trimLimit(out, limit), nil
}

func (s *MemoryStore) CountByStatus(status domain.Status) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, j := range s.jobs {
		if j.Status == status {
			n++
		}
	}
	return n, nil
}

func (s *MemoryStore) TryClaim(id string) (*domain.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok || j.Status != domain.StatusPending {
		return nil, false
	}
	j.Status = domain.StatusRunning
	j.UpdatedAt = time.Now().UTC()
	return cloneJob(j), true
}

// FileStore persists jobs as JSON files under Dir.
type FileStore struct {
	Dir string
}

func NewFileStore(dir string) *FileStore {
	return &FileStore{Dir: dir}
}

func (s *FileStore) jobPath(id string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, id)
	if safe == "" {
		safe = "job"
	}
	return filepath.Join(s.Dir, safe+".json")
}

func (s *FileStore) Put(j *domain.Job) error {
	if j == nil || j.ID == "" {
		return fmt.Errorf("invalid job")
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	return writeJobAtomic(s.jobPath(j.ID), j)
}

func (s *FileStore) Get(id string) (*domain.Job, bool) {
	j, err := readJobFile(s.jobPath(id))
	if err != nil {
		return nil, false
	}
	return j, true
}

func (s *FileStore) ListPending() ([]*domain.Job, error) {
	return s.ListByStatus(domain.StatusPending, 0)
}

func (s *FileStore) ListByStatus(status domain.Status, limit int) ([]*domain.Job, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*domain.Job
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		j, err := readJobFile(filepath.Join(s.Dir, e.Name()))
		if err != nil {
			continue
		}
		if status != "" && j.Status != status {
			continue
		}
		out = append(out, j)
	}
	sortJobsByCreated(out)
	return trimLimit(out, limit), nil
}

func (s *FileStore) CountByStatus(status domain.Status) (int, error) {
	list, err := s.ListByStatus(status, 0)
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

func (s *FileStore) TryClaim(id string) (*domain.Job, bool) {
	path := s.jobPath(id)
	j, err := readJobFile(path)
	if err != nil || j.Status != domain.StatusPending {
		return nil, false
	}
	j.Status = domain.StatusRunning
	j.UpdatedAt = time.Now().UTC()
	if err := writeJobAtomic(path, j); err != nil {
		return nil, false
	}
	return j, true
}

func sortJobsByCreated(jobs []*domain.Job) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})
}

func trimLimit(jobs []*domain.Job, limit int) []*domain.Job {
	if limit <= 0 || len(jobs) <= limit {
		return jobs
	}
	return jobs[:limit]
}

func writeJobAtomic(path string, j *domain.Job) error {
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readJobFile(path string) (*domain.Job, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var j domain.Job
	if err := json.Unmarshal(b, &j); err != nil {
		return nil, err
	}
	return &j, nil
}

func cloneJob(j *domain.Job) *domain.Job {
	if j == nil {
		return nil
	}
	cp := *j
	if j.Parameters != nil {
		cp.Parameters = make(map[string]string, len(j.Parameters))
		for k, v := range j.Parameters {
			cp.Parameters[k] = v
		}
	}
	return &cp
}
