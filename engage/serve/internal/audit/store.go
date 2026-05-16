package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event is a persisted tool run audit record.
type Event struct {
	Subject string    `json:"subject"`
	Tool    string    `json:"tool"`
	Target  string    `json:"target"`
	JobID   string    `json:"job_id"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
	At      time.Time `json:"at"`
}

// Store appends audit events as JSONL.
type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(dir string) (*Store, error) {
	if dir == "" {
		dir = "/var/veil/engage/audit"
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "events.jsonl")}, nil
}

func (s *Store) Append(e Event) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func (s *Store) Recent(limit int) ([]Event, error) {
	if s == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var lines []Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			lines = append(lines, e)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	// reverse chrono (newest first)
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines, nil
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// ExportNDJSON returns all events since the given time (zero = all), newest first in Recent order.
func (s *Store) ExportNDJSON(since time.Time) ([]byte, error) {
	events, err := s.readAllEvents()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	for _, e := range events {
		if !since.IsZero() && e.At.Before(since) {
			continue
		}
		b, err := json.Marshal(e)
		if err != nil {
			continue
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func (s *Store) readAllEvents() ([]Event, error) {
	if s == nil {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var lines []Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			lines = append(lines, e)
		}
	}
	return lines, sc.Err()
}

// OpenStore creates store or returns nil if dir empty and disabled.
func OpenStore(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("audit dir required")
	}
	return NewStore(dir)
}
