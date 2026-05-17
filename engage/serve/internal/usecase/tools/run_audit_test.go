package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

type auditSpy struct {
	calls int
	last  audit.Event
}

func (s *auditSpy) Append(e audit.Event) error {
	s.calls++
	s.last = e
	return nil
}

func TestRunOnce_emitsAuditWhenBinaryMissing(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{{
		Name: "missing_bin_scan", Category: toolid.CategoryWeb, Binary: "definitely-not-on-path-xyz",
		Enabled: true, TimeoutSec: 30,
	}})
	spy := &auditSpy{}
	r := &Runner{
		Registry: reg,
		Exec:     &runner.Executor{WorkDir: t.TempDir()},
		Audit:    audit.NewWithStore(slog.New(slog.NewTextHandler(os.Stderr, nil)), spy),
	}
	out := r.runOnce(context.Background(), "subj", "missing_bin_scan", contract.ToolRunRequest{Target: "example.com"})
	if out.Success {
		t.Fatal("expected failure when binary missing")
	}
	if spy.calls != 1 {
		t.Fatalf("audit append calls = %d, want 1 (events pipeline)", spy.calls)
	}
	if spy.last.Tool != "missing_bin_scan" || spy.last.Target != "example.com" {
		t.Fatalf("audit event = %+v", spy.last)
	}
}
