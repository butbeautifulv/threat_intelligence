package intelligence

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
)

func TestCreateAttackChain_usesPattern(t *testing.T) {
	reg := tools.NewRegistry([]tool.Spec{
		{Name: "nmap_scan", Binary: "nmap", Enabled: true},
		{Name: "httpx_probe", Binary: "httpx", Enabled: true},
		{Name: "nuclei_scan", Binary: "nuclei", Enabled: true},
	})
	s := &Service{Registry: reg, Engine: DefaultDecisionEngine()}
	chain := s.CreateAttackChain(context.Background(), "https://example.com", "comprehensive")
	if chain["pattern"] == "ranked_fallback" {
		t.Fatalf("expected named pattern, got fallback: %v", chain["pattern"])
	}
	steps, _ := chain["steps"].([]map[string]any)
	if len(steps) == 0 {
		t.Fatal("expected steps")
	}
}
