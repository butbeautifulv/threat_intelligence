package process

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Record tracks a running or finished subprocess.
type Record struct {
	PID       int        `json:"pid"`
	Tool      string     `json:"tool,omitempty"`
	Target    string     `json:"target,omitempty"`
	Command   string     `json:"command,omitempty"`
	Session   string     `json:"session,omitempty"` // docker exec session when PID is synthetic
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Manager tracks engage subprocesses for admin APIs.
type Manager struct {
	mu           sync.RWMutex
	records      map[int]*Record
	nextDockerID int64
}

func NewManager() *Manager {
	return &Manager{records: make(map[int]*Record)}
}

func (m *Manager) Register(pid int, tool, target, command string) {
	m.RegisterSession(pid, tool, target, command, "")
}

func (m *Manager) RegisterSession(pid int, tool, target, command, session string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[pid] = &Record{
		PID:       pid,
		Tool:      tool,
		Target:    target,
		Command:   command,
		Session:   session,
		Status:    "running",
		StartedAt: time.Now().UTC(),
	}
}

// RegisterDocker records a docker exec session with a unique negative PID.
func (m *Manager) RegisterDocker(tool, target, command, session string) int {
	pid := int(atomic.AddInt64(&m.nextDockerID, -1))
	m.RegisterSession(pid, tool, target, command, session)
	return pid
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

func (m *Manager) Pause(pid int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[pid]
	if !ok {
		return fmt.Errorf("pid %d not found", pid)
	}
	if r.Status != "running" {
		return fmt.Errorf("pid %d is not running", pid)
	}
	r.Status = "paused"
	return nil
}

func (m *Manager) Resume(pid int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[pid]
	if !ok {
		return fmt.Errorf("pid %d not found", pid)
	}
	if r.Status != "paused" {
		return fmt.Errorf("pid %d is not paused", pid)
	}
	r.Status = "running"
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
