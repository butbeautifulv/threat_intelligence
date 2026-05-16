package intelligence

import "testing"

func TestBuildTargetProfile_attackSurface(t *testing.T) {
	p := BuildTargetProfile("https://example.com", "web", []string{"php", "nginx"}, "wordpress", []string{"93.184.216.34"}, 0)
	if p.AttackSurfaceScore < 7 {
		t.Fatalf("attack surface: %v", p.AttackSurfaceScore)
	}
	if p.ConfidenceScore < 0.8 {
		t.Fatalf("confidence: %v", p.ConfidenceScore)
	}
	if p.RiskLevel == "" {
		t.Fatal("risk level empty")
	}
}

func TestCreateAttackChain_successProbability(t *testing.T) {
	s := &Service{
		Registry: testRegistry(webCatalogSpecs()),
		Engine:   DefaultDecisionEngine(),
	}
	chain := s.CreateAttackChain(t.Context(), "https://example.com", "recon")
	sp, ok := chain["success_probability"].(float64)
	if !ok || sp <= 0 {
		t.Fatalf("success_probability: %v", chain["success_probability"])
	}
	steps, ok := chain["steps"].([]map[string]any)
	if !ok || len(steps) == 0 {
		t.Fatalf("steps: %v", chain["steps"])
	}
	if _, ok := steps[0]["execution_time_estimate"]; !ok {
		t.Fatal("missing execution_time_estimate on step")
	}
}
