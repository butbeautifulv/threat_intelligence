package mcpserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tooldispatch"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

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
	d := tooldispatch.NewDispatcher(runner, &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}, cveSvc, nil, nil, nil, nil, nil, "", nil)
	srv := NewServerWithDispatch(d, runner, nil, slog.Default())

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
	content, _ := m["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected content")
	}
}

func TestCallIntelBridge_generateExploitFromCVE(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "generate_exploit_from_cve", Category: toolid.CategoryIntel, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	cveSvc := cve.NewService(nil, mockNVD{})
	d := tooldispatch.NewDispatcher(runner, &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}, cveSvc, nil, nil, nil, nil, nil, "", nil)
	srv := NewServerWithDispatch(d, runner, nil, slog.Default())

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
	content, _ := m["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected content")
	}
}

func TestCallIntelBridge_analyzeTarget(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "analyze_target_intelligence", Category: toolid.CategoryIntel, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	intel := &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}
	d := tooldispatch.NewDispatcher(runner, intel, nil, nil, nil, nil, nil, nil, "", nil)
	srv := NewServerWithDispatch(d, runner, nil, slog.Default())

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
	content, _ := m["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected content")
	}
}

func TestCallAgentTool_getProcessDashboard(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "get_process_dashboard", Category: toolid.CategoryWeb, Enabled: true},
	})
	runner := &toolsuc.Runner{Registry: reg}
	proc := process.NewManager()
	d := tooldispatch.NewDispatcher(runner, nil, nil, nil, nil, nil, proc, nil, "", nil)
	srv := NewServerWithDispatch(d, runner, nil, slog.Default())
	out, err := srv.callTool(context.Background(), "get_process_dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	content, _ := m["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected content")
	}
}
