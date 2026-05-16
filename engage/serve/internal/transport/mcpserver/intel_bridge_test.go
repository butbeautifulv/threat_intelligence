package mcpserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func TestIsIntelBridgeTool(t *testing.T) {
	if !IsIntelBridgeTool("comprehensive_api_audit", tool.Spec{Category: toolid.CategoryWeb}) {
		t.Fatal("comprehensive_api_audit should bridge")
	}
	if !IsIntelBridgeTool("analyze_target_intelligence", tool.Spec{Category: toolid.CategoryIntel}) {
		t.Fatal("intelligence category should bridge")
	}
	if !IsIntelBridgeTool("monitor_cve_feeds", tool.Spec{Category: toolid.CategoryWeb}) {
		t.Fatal("monitor_cve_feeds should bridge")
	}
}

type mockNVD struct{}

func (mockNVD) FetchCVE(_ context.Context, cveID string) (*cve.CVEEntry, error) {
	return &cve.CVEEntry{
		CVEID:       cveID,
		Description: "SQL injection in login form",
		Severity:    "HIGH",
		CVSSScore:   8.1,
	}, nil
}

func (mockNVD) FetchRecent(_ context.Context, _ int, _ string) ([]cve.CVEEntry, error) {
	return []cve.CVEEntry{
		{CVEID: "CVE-2020-0001", Description: "xss reflected", Severity: "HIGH", CVSSScore: 7.5},
	}, nil
}

func TestCallIntelBridge_monitorCVE(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "monitor_cve_feeds", Category: toolid.CategoryIntel, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	cveSvc := cve.NewService(nil, mockNVD{})
	srv := NewServerFull(runner, &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}, cveSvc, nil, nil, nil, nil, nil, nil, slog.Default(), "", nil)

	out, err := srv.callTool(context.Background(), "monitor_cve_feeds", map[string]any{
		"hours": 24,
	})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	if len(m["content"].([]map[string]any)) == 0 {
		t.Fatal("expected content")
	}
}

func TestCallIntelBridge_generateExploitFromCVE(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "generate_exploit_from_cve", Category: toolid.CategoryIntel, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	cveSvc := cve.NewService(nil, mockNVD{})
	srv := NewServerFull(runner, &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}, cveSvc, nil, nil, nil, nil, nil, nil, slog.Default(), "", nil)

	out, err := srv.callTool(context.Background(), "generate_exploit_from_cve", map[string]any{
		"cve_id":       "CVE-2020-0001",
		"exploit_type": "poc",
	})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	if len(m["content"].([]map[string]any)) == 0 {
		t.Fatal("expected content")
	}
}

func TestTryAgentTool_getProcessDashboard(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "get_process_dashboard", Category: toolid.CategoryWeb, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	proc := process.NewManager()
	srv := NewServerFull(runner, nil, nil, nil, nil, nil, proc, nil, nil, slog.Default(), "", nil)
	out, ok, err := srv.tryAgentTool(context.Background(), "get_process_dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected handled")
	}
	m, _ := out.(map[string]any)
	if len(m["content"].([]map[string]any)) == 0 {
		t.Fatal("expected content")
	}
}

func TestCallIntelBridge_analyzeTarget(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "analyze_target_intelligence", Category: toolid.CategoryIntel, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	intel := &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}
	srv := NewServerWithIntel(runner, intel, nil, nil, slog.Default(), "", nil)

	out, err := srv.callTool(context.Background(), "analyze_target_intelligence", map[string]any{
		"target": "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	if len(m["content"].([]map[string]any)) == 0 {
		t.Fatal("expected content")
	}
}
