package command

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func TestRunner_allowlist(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "nmap_scan", Category: toolid.CategoryNetwork, Binary: "echo", Enabled: true},
	})
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg)
	out := r.Run(context.Background(), "echo hello", false, nil)
	if out["success"] != true {
		t.Fatalf("out: %v", out)
	}
}

func TestRunner_rejectsUnknownBinary(t *testing.T) {
	reg := tools.NewRegistry(nil)
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg)
	out := r.Run(context.Background(), "/bin/false", false, nil)
	if out["success"] != false {
		t.Fatal("expected failure")
	}
}
