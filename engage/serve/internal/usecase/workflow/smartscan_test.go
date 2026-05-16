package workflow

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func echoSpecs() []tool.Spec {
	names := []string{"nuclei_scan", "httpx_probe", "gobuster_scan"}
	out := make([]tool.Spec, len(names))
	for i, n := range names {
		out[i] = tool.Spec{
			Name: n, Category: toolid.CategoryWeb, Binary: "echo",
			Enabled: true, ArgsTemplate: []string{"{target}"}, TimeoutSec: 5,
		}
	}
	return out
}

func TestSmartScan_syncRespectsMaxTools(t *testing.T) {
	reg := tools.NewRegistry(echoSpecs())
	intel := &intelligence.Service{Registry: reg, Engine: intelligence.DefaultDecisionEngine()}
	exec := &runner.Executor{WorkDir: t.TempDir()}
	tr := &toolsuc.Runner{Registry: reg, Exec: exec}
	s := &Service{Intel: intel, Tools: tr}
	out := s.SmartScan(context.Background(), "", SmartScanRequest{
		Target: "https://example.com", MaxTools: 2, Async: false,
	})
	selected, _ := out["tools_selected"].([]string)
	if len(selected) > 2 {
		t.Fatalf("expected at most 2 tools, got %v", selected)
	}
	executed, _ := out["tools_executed"].([]map[string]any)
	if len(executed) > 2 {
		t.Fatalf("expected at most 2 executions, got %d", len(executed))
	}
}

func TestSmartScan_asyncEnqueuesJobs(t *testing.T) {
	reg := tools.NewRegistry(echoSpecs())
	intel := &intelligence.Service{Registry: reg, Engine: intelligence.DefaultDecisionEngine()}
	tr := &toolsuc.Runner{Registry: reg, Exec: &runner.Executor{WorkDir: t.TempDir()}}
	jobs := jobuc.NewQueue(tr, jobuc.WithMode(jobuc.ModeMemory))
	s := &Service{Intel: intel, Tools: tr, Jobs: jobs}
	out := s.SmartScan(context.Background(), "tester", SmartScanRequest{
		Target: "https://example.com", MaxTools: 2, Async: true,
	})
	if out["status"] != "queued" {
		t.Fatalf("status: %v", out["status"])
	}
	executed, _ := out["tools_executed"].([]map[string]any)
	if len(executed) == 0 {
		t.Fatal("expected queued tools")
	}
}
