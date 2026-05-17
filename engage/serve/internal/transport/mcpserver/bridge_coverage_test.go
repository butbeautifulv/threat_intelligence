package mcpserver

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cache"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func bridgeAPISpecs() []tool.Spec {
	names := []string{
		"clear_cache", "get_cache_stats", "list_files", "http_repeater",
		"api_fuzzer", "generate_payload", "kube_hunter_scan",
	}
	specs := make([]tool.Spec, 0, len(names))
	for _, n := range names {
		specs = append(specs, tool.Spec{Name: n, Category: toolid.CategoryWeb, Enabled: true, Binary: bridgeBinaryForTest(n)})
	}
	return specs
}

func bridgeBinaryForTest(name string) string {
	switch {
	case strings.HasPrefix(name, "get_"), strings.HasPrefix(name, "list_"):
		return "get"
	case strings.HasPrefix(name, "http_"):
		return "http"
	case strings.HasPrefix(name, "kube_"):
		return "kube"
	case name == "clear_cache":
		return "clear"
	case name == "api_fuzzer":
		return "api"
	case name == "generate_payload":
		return "generate"
	default:
		return "advanced"
	}
}

func TestBridgeWorkflowTools_structuredSuccess(t *testing.T) {
	reg := tools.NewRegistry(bridgeAPISpecs())
	runner := &toolsuc.Runner{Registry: reg, Cache: cache.New(0)}
	fileMgr, err := files.NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	proc := process.NewManager()
	intel := &intelligence.Service{Engine: intelligence.DefaultDecisionEngine()}
	srv := NewServerWithIntel(runner, intel, nil, nil, slog.Default(), "", fileMgr)
	srv.processes = proc

	for _, name := range []string{"clear_cache", "get_cache_stats", "list_files", "http_repeater", "generate_payload"} {
		t.Run(name, func(t *testing.T) {
			out, err := srv.callTool(context.Background(), name, map[string]any{"target": "https://example.com"})
			if err != nil {
				t.Fatal(err)
			}
			text := bridgeResultText(t, out)
			if strings.Contains(text, "not mapped") {
				t.Fatalf("unmapped: %s", text)
			}
			if !strings.Contains(text, `"success":true`) && !strings.Contains(text, `"success": true`) {
				t.Fatalf("expected success in %s", text)
			}
		})
	}
}

func bridgeResultText(t *testing.T, out any) string {
	t.Helper()
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	content, ok := m["content"].([]map[string]any)
	if !ok || len(content) == 0 {
		t.Fatal("expected content")
	}
	return content[0]["text"].(string)
}
