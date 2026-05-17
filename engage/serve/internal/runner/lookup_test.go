package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookupBinary_burpsuiteCandidates(t *testing.T) {
	dir := t.TempDir()
	burp := filepath.Join(dir, "burpsuite_scan")
	if err := os.WriteFile(burp, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	got, err := LookupBinary("burpsuite")
	if err != nil {
		t.Fatalf("LookupBinary(burpsuite): %v", err)
	}
	if got != burp {
		t.Fatalf("got %q want %q", got, burp)
	}
}
