package visual

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ToolProgress tracks one tool in a scan.
type ToolProgress struct {
	Tool          string  `json:"tool"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	ExecutionTime float64 `json:"execution_time,omitempty"`
}

// ScanProgress is pollable scan/chain state.
type ScanProgress struct {
	ScanID           string         `json:"scan_id"`
	Target           string         `json:"target"`
	ScanType         string         `json:"scan_type,omitempty"`
	Status           string         `json:"status"`
	ProgressPercent  float64        `json:"progress_percent"`
	ToolsTotal       int            `json:"tools_total"`
	ToolsCompleted   int            `json:"tools_completed"`
	CurrentTool      string         `json:"current_tool,omitempty"`
	Tools            []ToolProgress `json:"tools"`
	StartedAt        time.Time      `json:"started_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// Store holds in-memory scan progress.
type Store struct {
	mu    sync.RWMutex
	scans map[string]*ScanProgress
}

func NewStore() *Store {
	return &Store{scans: make(map[string]*ScanProgress)}
}

// Create starts tracking a new scan.
func (s *Store) Create(target, scanType string, tools []string) string {
	id := newScanID()
	now := time.Now().UTC()
	toolRows := make([]ToolProgress, 0, len(tools))
	for _, t := range tools {
		toolRows = append(toolRows, ToolProgress{Tool: t, Status: "pending", Progress: 0})
	}
	s.mu.Lock()
	s.scans[id] = &ScanProgress{
		ScanID:          id,
		Target:          target,
		ScanType:        scanType,
		Status:          "running",
		ToolsTotal:      len(tools),
		Tools:           toolRows,
		StartedAt:       now,
		UpdatedAt:       now,
	}
	s.mu.Unlock()
	return id
}

// StartTool marks a tool as running.
func (s *Store) StartTool(scanID, tool string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sp, ok := s.scans[scanID]
	if !ok {
		return
	}
	sp.CurrentTool = tool
	sp.UpdatedAt = time.Now().UTC()
	for i := range sp.Tools {
		if sp.Tools[i].Tool == tool {
			sp.Tools[i].Status = "running"
			sp.Tools[i].Progress = 0.1
		}
	}
	s.recalc(sp)
}

// CompleteTool marks a tool done or failed.
func (s *Store) CompleteTool(scanID, tool, status string, execSeconds float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sp, ok := s.scans[scanID]
	if !ok {
		return
	}
	for i := range sp.Tools {
		if sp.Tools[i].Tool == tool {
			sp.Tools[i].Status = status
			sp.Tools[i].Progress = 1
			sp.Tools[i].ExecutionTime = execSeconds
		}
	}
	sp.UpdatedAt = time.Now().UTC()
	s.recalc(sp)
}

// Finish marks the scan completed or failed.
func (s *Store) Finish(scanID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sp, ok := s.scans[scanID]; ok {
		sp.Status = status
		sp.ProgressPercent = 100
		if status == "completed" {
			sp.ToolsCompleted = sp.ToolsTotal
		}
		sp.UpdatedAt = time.Now().UTC()
	}
}

// Get returns a copy of scan progress.
func (s *Store) Get(scanID string) (ScanProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sp, ok := s.scans[scanID]
	if !ok {
		return ScanProgress{}, false
	}
	cp := *sp
	cp.Tools = append([]ToolProgress(nil), sp.Tools...)
	return cp, true
}

func newScanID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("scan-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

func (s *Store) recalc(sp *ScanProgress) {
	done := 0
	var sum float64
	for _, t := range sp.Tools {
		sum += t.Progress
		if t.Status == "success" || t.Status == "failed" || t.Status == "done" {
			done++
		}
	}
	sp.ToolsCompleted = done
	if sp.ToolsTotal > 0 {
		sp.ProgressPercent = (sum / float64(sp.ToolsTotal)) * 100
	}
}
