package process

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// Record tracks a running or finished subprocess.
type Record struct {
	PID       int        `json:"pid"`
	Tool      string     `json:"tool,omitempty"`
	Target    string     `json:"target,omitempty"`
	Command   string     `json:"command,omitempty"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Manager tracks engage subprocesses for admin APIs.
type Manager struct {
	mu      sync.RWMutex
	records map[int]*Record
}

func NewManager() *Manager {
	return &Manager{records: make(map[int]*Record)}
}

func (m *Manager) Register(pid int, tool, target, command string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[pid] = &Record{
		PID:       pid,
		Tool:      tool,
		Target:    target,
		Command:   command,
		Status:    "running",
		StartedAt: time.Now().UTC(),
	}
}

func (m *Manager) Finish(pid int, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.records[pid]; ok {
		r.Status = status
		now := time.Now().UTC()
		r.EndedAt = &now
	}
}

func (m *Manager) List() []Record {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Record, 0, len(m.records))
	for _, r := range m.records {
		out = append(out, *r)
	}
	return out
}

func (m *Manager) Get(pid int) (*Record, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[pid]
	if !ok {
		return nil, false
	}
	cp := *r
	return &cp, true
}

func (m *Manager) Terminate(_ context.Context, pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("terminate pid %d: %w", pid, err)
		}
	}
	m.Finish(pid, "terminated")
	return nil
}

func (m *Manager) Dashboard() map[string]any {
	list := m.List()
	running := 0
	for _, r := range list {
		if r.Status == "running" {
			running++
		}
	}
	return map[string]any{
		"total":     len(list),
		"running":   running,
		"processes": list,
	}
}
