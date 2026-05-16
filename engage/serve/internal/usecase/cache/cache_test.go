package cache

import (
	"testing"
	"time"
)

func TestStore_setGetClear(t *testing.T) {
	s := New(time.Minute)
	s.Set("k", "v")
	if got, ok := s.Get("k"); !ok || got != "v" {
		t.Fatalf("get: %q ok=%v", got, ok)
	}
	if n := s.Clear(); n != 1 {
		t.Fatalf("cleared %d", n)
	}
	if _, ok := s.Get("k"); ok {
		t.Fatal("expected miss after clear")
	}
}
