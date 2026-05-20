package cataloglink

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func testRepoRoot(t *testing.T) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Skip("repo root not found")
	}
	t.Setenv("VEIL_REPO_ROOT", root)
	catalogOnce = sync.Once{}
	catalogNames = nil
}

func TestResolveMentions_skipsShortFalsePositives(t *testing.T) {
	testRepoRoot(t)
	got := ResolveMentions([]string{"dd", "dcfldd"})
	if len(got) != 0 {
		t.Fatalf("dd/dcfldd should not map to catalog params: %v", got)
	}
}

func TestResolveMentions_nmapAlias(t *testing.T) {
	testRepoRoot(t)
	got := ResolveMentions([]string{"nmap"})
	if len(got) != 1 || got[0] != "nmap_scan" {
		t.Fatalf("nmap: got %v", got)
	}
}
