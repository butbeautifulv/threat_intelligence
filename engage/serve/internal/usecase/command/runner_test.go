package command

import (
	"context"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func TestRunner_rejectsShellMetachar(t *testing.T) {
	reg := tools.NewRegistry(nil)
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg, false)
	out := r.Run(context.Background(), "id; rm -rf /", false, nil)
	if out["success"] != false {
		t.Fatalf("expected failure: %v", out)
	}
	errMsg, _ := out["error"].(string)
	if !strings.Contains(errMsg, "metachar") {
		t.Fatalf("expected metachar error: %v", out)
	}
}

func TestRunner_allowlist(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "nmap_scan", Category: toolid.CategoryNetwork, Binary: "echo", Enabled: true},
	})
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg, false)
	out := r.Run(context.Background(), "echo hello", false, nil)
	if out["success"] != true {
		t.Fatalf("out: %v", out)
	}
}

func TestRunner_rejectsUnknownBinary(t *testing.T) {
	reg := tools.NewRegistry(nil)
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg, false)
	out := r.Run(context.Background(), "/bin/false", false, nil)
	if out["success"] != false {
		t.Fatal("expected failure")
	}
}

func TestRunner_denyRawIgnoresEnv(t *testing.T) {
	t.Setenv("ENGAGE_ALLOW_RAW_COMMAND", "1")
	reg := tools.NewRegistry(nil)
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg, false)
	_, _, err := r.parseCommand("id")
	if err == nil {
		t.Fatal("expected allowlist error when allowRaw=false")
	}
}

func TestRunner_allowRawPermitsAnyBinary(t *testing.T) {
	reg := tools.NewRegistry(nil)
	r := New(&runner.Executor{WorkDir: t.TempDir()}, reg, true)
	bin, _, err := r.parseCommand("id")
	if err != nil || bin != "id" {
		t.Fatalf("parse: bin=%q err=%v", bin, err)
	}
}
