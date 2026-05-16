package process

import (
	"context"
	"fmt"
	"os"
	goruntime "runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Record tracks a running or finished subprocess.
type Record struct {
	PID            int        `json:"pid"`
	Tool           string     `json:"tool,omitempty"`
	Target         string     `json:"target,omitempty"`
	Command        string     `json:"command,omitempty"`
	Session        string     `json:"session,omitempty"`
	Status         string     `json:"status"`
	Progress       float64    `json:"progress"`
	LastOutput     string     `json:"last_output,omitempty"`
	BytesProcessed int64      `json:"bytes_processed,omitempty"`
	ETA            float64    `json:"eta_seconds,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
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
		Progress:  0,
		StartedAt: time.Now().UTC(),
	}
}

func (m *Manager) RegisterDocker(tool, target, command, session string) int {
	pid := int(atomic.AddInt64(&m.nextDockerID, -1))
	m.RegisterSession(pid, tool, target, command, session)
	return pid
}

func (m *Manager) UpdateProgress(pid int, progress float64, lastOutput string, bytesProcessed int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[pid]
	if !ok {
		return
	}
	if progress >= 0 {
		r.Progress = progress
		if progress > 0 && progress < 1 {
			elapsed := time.Since(r.StartedAt).Seconds()
			r.ETA = (elapsed / progress) * (1 - progress)
		}
	}
	if lastOutput != "" {
		if len(lastOutput) > 200 {
			lastOutput = lastOutput[len(lastOutput)-200:]
		}
		r.LastOutput = lastOutput
	}
	if bytesProcessed > 0 {
		r.BytesProcessed = bytesProcessed
	}
}

func (m *Manager) Finish(pid int, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.records[pid]; ok {
		r.Status = status
		r.Progress = 1
		r.ETA = 0
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
	processes := make([]map[string]any, 0, len(list))
	now := time.Now().UTC()
	for _, r := range list {
		if r.Status == "running" {
			running++
		}
		runtimeSec := now.Sub(r.StartedAt).Seconds()
		if r.EndedAt != nil {
			runtimeSec = r.EndedAt.Sub(r.StartedAt).Seconds()
		}
		processes = append(processes, map[string]any{
			"pid":                 r.PID,
			"tool":                r.Tool,
			"target":              r.Target,
			"command":             r.Command,
			"status":              r.Status,
			"runtime":             fmt.Sprintf("%.1fs", runtimeSec),
			"progress_percent":    fmt.Sprintf("%.1f%%", r.Progress*100),
			"progress_fraction":   r.Progress,
			"last_output":         r.LastOutput,
			"bytes_processed":     r.BytesProcessed,
			"eta_seconds":         r.ETA,
		})
	}
	var mem goruntime.MemStats
	goruntime.ReadMemStats(&mem)
	return map[string]any{
		"timestamp":       now.Format(time.RFC3339),
		"total_processes": len(list),
		"total":           len(list),
		"running":         running,
		"system_load": map[string]any{
			"goroutines":      goruntime.NumGoroutine(),
			"memory_alloc_mb": float64(mem.Alloc) / (1024 * 1024),
		},
		"processes": processes,
	}
}
