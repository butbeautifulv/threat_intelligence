package config

import (
	"os"
	"testing"
)

func TestMaxParallel_envAndCap(t *testing.T) {
	t.Setenv("ENGAGE_MAX_PARALLEL", "99")
	if got := maxParallel(); got != 32 {
		t.Fatalf("cap got %d", got)
	}
	t.Setenv("ENGAGE_MAX_PARALLEL", "2")
	if got := maxParallel(); got != 2 {
		t.Fatalf("got %d", got)
	}
}

func TestLoadAPI_maxParallelDefault(t *testing.T) {
	os.Unsetenv("ENGAGE_MAX_PARALLEL")
	cfg := LoadAPI()
	if cfg.MaxParallel != 5 {
		t.Fatalf("MaxParallel %d", cfg.MaxParallel)
	}
}
