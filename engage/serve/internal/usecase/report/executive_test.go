package report

import (
	"testing"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
)

func TestBuildExecutiveSummary(t *testing.T) {
	findings := []domain.Finding{
		{Title: "SQLi", Severity: domain.SeverityHigh},
		{Title: "Info leak", Severity: domain.SeverityLow},
	}
	scan := map[string]any{
		"tools_executed": []map[string]any{
			{"tool": "nmap", "execution_time": 2.0},
		},
	}
	es := BuildExecutiveSummary("https://example.com", scan, findings, "medium", []string{"nginx"})
	if es.TotalFindings != 2 {
		t.Fatalf("total %d", es.TotalFindings)
	}
	if es.High != 1 {
		t.Fatalf("high %d", es.High)
	}
	if es.RiskPosture != "high" {
		t.Fatalf("posture %s", es.RiskPosture)
	}
	if len(es.TopRisks) == 0 {
		t.Fatal("expected top risks")
	}
	if len(es.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
}
