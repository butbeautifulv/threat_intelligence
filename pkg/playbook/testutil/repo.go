// Package testutil helps playbook tests locate the Veil repo root via VEIL_REPO_ROOT.
package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// RepoRoot returns the Veil monorepo root (directory containing versions.env).
func RepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "versions.env")); err != nil {
		t.Skip("repo root not found (versions.env)")
	}
	return root
}

// SetRepoRoot sets VEIL_REPO_ROOT for playbook loaders.
func SetRepoRoot(t *testing.T) string {
	t.Helper()
	root := RepoRoot(t)
	t.Setenv("VEIL_REPO_ROOT", root)
	return root
}
