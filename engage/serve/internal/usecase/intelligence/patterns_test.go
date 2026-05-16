package intelligence

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
)

func TestAttackPatterns_count(t *testing.T) {
	p := AttackPatterns()
	if len(p) < 20 {
		t.Fatalf("expected >= 20 patterns, got %d", len(p))
	}
}

func TestAttackPatterns_eachHasSteps(t *testing.T) {
	for key, steps := range AttackPatterns() {
		if len(steps) == 0 {
			t.Fatalf("pattern %q has no steps", key)
		}
	}
}

func TestSelectPatternKey_binaryAndCloud(t *testing.T) {
	if got := SelectPatternKey("binary", "exploit"); got != "binary_exploitation" {
		t.Fatalf("binary: %q", got)
	}
	if got := SelectPatternKey("cloud", "multi-cloud"); got != "multi_cloud_assessment" {
		t.Fatalf("multi-cloud: %q", got)
	}
}

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
