package mcpserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
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
