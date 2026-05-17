package intelligence

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
)

func TestCreateAttackChain_webRecon_hasParams(t *testing.T) {
	pattern := AttackPatterns()["web_reconnaissance"]
	if len(pattern) == 0 || pattern[0].Params["ports"] == "" {
		t.Fatal("web_reconnaissance should include nmap ports param")
	}
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "nmap", Binary: "nmap", Enabled: true},
		{Name: "httpx", Binary: "httpx", Enabled: true},
	})
	s := &Service{Registry: reg, Engine: DefaultDecisionEngine()}
	chain := s.CreateAttackChain(context.Background(), "https://example.com", "web")
	steps, _ := chain["steps"].([]map[string]any)
	for _, step := range steps {
		if step["tool"] == "nmap" {
			params, _ := step["parameters"].(map[string]string)
			if params != nil && params["ports"] != "" {
				return
			}
		}
	}
	t.Fatalf("expected nmap step with ports param, got %v", steps)
}
