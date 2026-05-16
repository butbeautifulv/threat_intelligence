package intelligence

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
)

type stubCVEIntel struct {
	details []map[string]any
	ids     []string
	paths   []map[string]any
}

func (s *stubCVEIntel) EnrichCorrelation(_ context.Context, indicators string, related []string) ([]map[string]any, []string) {
	_ = indicators
	_ = related
	return s.details, s.ids
}

func (s *stubCVEIntel) BuildCVEAttackPaths(_ context.Context, cveIDs []string) []map[string]any {
	_ = cveIDs
	return s.paths
}

func TestCorrelateThreatIntelligence_cveDetails(t *testing.T) {
	svc := &Service{
		Engine: DefaultDecisionEngine(),
		CVE: &stubCVEIntel{
			ids: []string{"CVE-2021-44228"},
			details: []map[string]any{
				{
					"success": true,
					"cve": map[string]any{
						"cve_id": "CVE-2021-44228",
						"severity": "CRITICAL",
					},
					"analysis": map[string]any{
						"exploitability_score":  0.9,
						"exploitability_level":  "CRITICAL",
						"vulnerability_type":    "rce",
					},
				},
			},
		},
	}
	out := svc.CorrelateThreatIntelligence(context.Background(), "https://example.com", "CVE-2021-44228")
	details, ok := out["cve_details"].([]map[string]any)
	if !ok || len(details) == 0 {
		t.Fatalf("missing cve_details: %#v", out["cve_details"])
	}
	if details[0]["cve_id"] != "CVE-2021-44228" {
		t.Fatalf("cve_id %v", details[0]["cve_id"])
	}
}

func TestDiscoverAttackChains_cvePaths(t *testing.T) {
	svc := &Service{
		Engine:   DefaultDecisionEngine(),
		Registry: tools.NewRegistry(nil),
		CVE: &stubCVEIntel{
			paths: []map[string]any{
				{
					"cve_id":                     "CVE-2021-44228",
					"severity":                   "CRITICAL",
					"exploitability_score":       0.95,
					"exploit_template_available": true,
				},
			},
		},
	}
	out := svc.DiscoverAttackChains(context.Background(), "CVE-2021-44228", "comprehensive")
	paths, ok := out["cve_attack_paths"].([]map[string]any)
	if !ok || len(paths) == 0 {
		t.Fatalf("missing cve_attack_paths: %#v", out["cve_attack_paths"])
	}
	stages, ok := out["cve_stages"].([]map[string]any)
	if !ok || len(stages) == 0 {
		t.Fatal("missing cve_stages")
	}
	if stages[0]["exploit_available"] != true {
		t.Fatalf("stage %#v", stages[0])
	}
}
