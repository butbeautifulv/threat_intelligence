package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt time.Time
}

// Store is a thread-safe TTL cache for tool/command output.
type Store struct {
	mu      sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
}

func New(ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &Store{entries: make(map[string]entry), ttl: ttl}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.value, true
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = entry{value: value, expiresAt: time.Now().Add(s.ttl)}
}

func (s *Store) Clear() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.entries)
	s.entries = make(map[string]entry)
	return n
}

func (s *Store) Stats() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	active := 0
	now := time.Now()
	for _, e := range s.entries {
		if now.Before(e.expiresAt) {
			active++
		}
	}
	return map[string]any{"entries": active, "ttl_sec": int(s.ttl.Seconds())}
}
