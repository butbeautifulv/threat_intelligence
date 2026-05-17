package intelligence

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

func testRegistry(specs []tool.Spec) *tools.Registry {
	return tools.NewRegistry(specs)
}

func webCatalogSpecs() []tool.Spec {
	names := []string{"httpx_probe", "nuclei_scan", "gobuster_scan", "nikto_scan", "ffuf_scan", "feroxbuster_scan", "sqlmap_scan"}
	out := make([]tool.Spec, len(names))
	for i, n := range names {
		out[i] = tool.Spec{Name: n, Category: toolid.CategoryWeb, Binary: "echo", Enabled: true}
	}
	return out
}

func TestSelectTools_web_ranksNucleiFirst(t *testing.T) {
	s := &Service{
		Registry: testRegistry(webCatalogSpecs()),
		Engine:   DefaultDecisionEngine(),
	}
	got := s.SelectTools(context.Background(), "web", "")
	if len(got) == 0 {
		t.Fatal("expected tools")
	}
	if got[0] != "nuclei_scan" {
		t.Fatalf("expected nuclei_scan first, got %v", got)
	}
}

func TestSelectTools_skipsDisabled(t *testing.T) {
	specs := webCatalogSpecs()
	for i := range specs {
		if specs[i].Name == "nuclei_scan" {
			specs[i].Enabled = false
		}
	}
	s := &Service{Registry: testRegistry(specs), Engine: DefaultDecisionEngine()}
	got := s.SelectTools(context.Background(), "web", "")
	for _, name := range got {
		if name == "nuclei_scan" {
			t.Fatalf("nuclei_scan should be excluded when disabled: %v", got)
		}
	}
}

func TestSelectTools_unknownUsesEngine(t *testing.T) {
	specs := []tool.Spec{
		{Name: "nmap_scan", Category: toolid.CategoryNetwork, Binary: "nmap", Enabled: true},
		{Name: "httpx_probe", Category: toolid.CategoryWeb, Binary: "httpx", Enabled: true},
		{Name: "subfinder_scan", Category: toolid.CategoryOSINT, Binary: "subfinder", Enabled: true},
		{Name: "nuclei_scan", Category: toolid.CategoryWeb, Binary: "nuclei", Enabled: true},
	}
	s := &Service{Registry: testRegistry(specs), Engine: DefaultDecisionEngine()}
	got := s.SelectTools(context.Background(), "unknown", "")
	if len(got) == 0 {
		t.Fatal("expected tools for unknown target type")
	}
	found := false
	for _, name := range got {
		if name == "subfinder_scan" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected subfinder_scan in %v", got)
	}
}

func TestSelectToolsForTarget_wordpressWpscanBoost(t *testing.T) {
	specs := webCatalogSpecs()
	specs = append(specs, tool.Spec{Name: "wpscan_analyze", Category: toolid.CategoryWeb, Binary: "wpscan", Enabled: true})
	s := &Service{Registry: testRegistry(specs), Engine: DefaultDecisionEngine()}
	got := s.SelectToolsForTarget(context.Background(), "web", "", "https://example.com/wp-admin")
	if len(got) == 0 {
		t.Fatal("expected tools")
	}
	if got[0] != "wpscan_analyze" {
		t.Fatalf("expected wpscan_analyze first for wordpress path, got %v", got)
	}
}

func TestCreateAttackChain_stepsOrdered(t *testing.T) {
	s := &Service{
		Registry: testRegistry(webCatalogSpecs()),
		Engine:   DefaultDecisionEngine(),
	}
	chain := s.CreateAttackChain(context.Background(), "https://example.com", "assess")
	steps, ok := chain["steps"].([]map[string]any)
	if !ok || len(steps) == 0 {
		t.Fatalf("steps: %v", chain["steps"])
	}
	if _, ok := chain["pattern"]; !ok {
		t.Fatalf("missing pattern key: %v", chain)
	}
}
