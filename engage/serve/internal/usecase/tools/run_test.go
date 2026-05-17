package tools

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/security"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func TestMergeParameters_targetAliasesURL(t *testing.T) {
	spec := tool.Spec{
		Name: "gobuster_scan",
		Parameters: []tool.Param{
			{Name: "target", Required: true},
			{Name: "url", Required: true},
			{Name: "mode", Default: "dir"},
		},
	}
	out := mergeParameters(spec, contract.ToolRunRequest{
		Target: "http://example.com",
		Parameters: map[string]string{
			"mode": "dir",
		},
	})
	if out["target"] != "http://example.com" {
		t.Fatalf("target = %q", out["target"])
	}
	if out["url"] != "http://example.com" {
		t.Fatalf("url alias = %q", out["url"])
	}
	if out["mode"] != "dir" {
		t.Fatalf("mode = %q", out["mode"])
	}
}

func TestMergeParameters_targetAliasesDomain(t *testing.T) {
	spec := tool.Spec{
		Name: "amass_scan",
		Parameters: []tool.Param{
			{Name: "target", Required: true},
			{Name: "domain", Required: true},
		},
	}
	out := mergeParameters(spec, contract.ToolRunRequest{Target: "example.com"})
	if out["domain"] != "example.com" {
		t.Fatalf("domain alias = %q", out["domain"])
	}
}

func TestMergeParameters_explicitURLNotOverwritten(t *testing.T) {
	spec := tool.Spec{
		Parameters: []tool.Param{
			{Name: "target", Required: true},
			{Name: "url", Required: true},
		},
	}
	out := mergeParameters(spec, contract.ToolRunRequest{
		Target: "http://fallback.com",
		Parameters: map[string]string{
			"url": "http://explicit.com",
		},
	})
	if out["url"] != "http://explicit.com" {
		t.Fatalf("url = %q, want explicit", out["url"])
	}
}

func TestRunOnce_targetGuardBlocksMetadata_beforeCatalogLookup(t *testing.T) {
	const metadataTarget = "http://169.254.169.254/latest/meta-data/"
	cases := []struct {
		name     string
		toolName string
		specs    []tool.Spec
	}{
		{
			name:     "unknown_tool",
			toolName: "not_in_catalog_scan",
			specs:    nil,
		},
		{
			name:     "disabled_tool",
			toolName: "disabled_scan",
			specs: []tool.Spec{{
				Name: "disabled_scan", Category: toolid.CategoryWeb, Binary: "echo",
				Enabled: false, TimeoutSec: 30,
			}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := tools.NewRegistry(tc.specs)
			spy := &auditSpy{}
			r := &Runner{
				Registry:    reg,
				Exec:        &runner.Executor{WorkDir: t.TempDir()},
				Audit:       audit.NewWithStore(slog.New(slog.NewTextHandler(os.Stderr, nil)), spy),
				TargetGuard: security.TargetGuardBlock,
			}
			out := r.runOnce(context.Background(), "subj", tc.toolName, contract.ToolRunRequest{Target: metadataTarget})
			if out.Success {
				t.Fatal("expected blocked run")
			}
			if !strings.Contains(out.Error, "target blocked by ENGAGE_TARGET_GUARD") {
				t.Fatalf("error = %q, want target guard block before catalog", out.Error)
			}
			if strings.Contains(out.Error, "unknown tool") || strings.Contains(out.Error, "tool disabled") {
				t.Fatalf("catalog lookup ran before guard: %q", out.Error)
			}
			if spy.calls != 1 {
				t.Fatalf("audit calls = %d, want 1 for guard block", spy.calls)
			}
		})
	}
}

func TestRunOnce_targetGuardOff_stillBlocksMetadataDenylist(t *testing.T) {
	reg := tools.NewRegistry(nil)
	r := &Runner{
		Registry:    reg,
		Exec:        &runner.Executor{WorkDir: t.TempDir()},
		TargetGuard: security.TargetGuardOff,
	}
	out := r.runOnce(context.Background(), "", "missing_scan", contract.ToolRunRequest{
		Target: "http://169.254.169.254/latest/meta-data/",
	})
	if out.Success {
		t.Fatal("expected failure")
	}
	if !strings.Contains(out.Error, "target blocked by ENGAGE_TARGET_GUARD") {
		t.Fatalf("denylist must block even when guard off: %q", out.Error)
	}
	if strings.Contains(out.Error, "unknown tool") {
		t.Fatalf("catalog lookup ran before denylist guard: %q", out.Error)
	}
}

func TestRunOnce_targetGuardOff_allowsRFC1918(t *testing.T) {
	reg := tools.NewRegistry(nil)
	r := &Runner{
		Registry:    reg,
		Exec:        &runner.Executor{WorkDir: t.TempDir()},
		TargetGuard: security.TargetGuardOff,
	}
	out := r.runOnce(context.Background(), "", "missing_scan", contract.ToolRunRequest{
		Target: "10.0.0.1",
	})
	if strings.Contains(out.Error, "target blocked") {
		t.Fatalf("RFC1918 should be allowed when guard off: %q", out.Error)
	}
	if !strings.Contains(out.Error, "unknown tool") {
		t.Fatalf("expected catalog error, got %q", out.Error)
	}
}
