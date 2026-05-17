package report

import (
	"testing"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func TestFromSmartScan_includesFindings(t *testing.T) {
	scan := map[string]any{
		"status":    "completed",
		"objective": "quick",
		"findings": []any{
			map[string]any{
				"title": "test", "severity": "high", "description": "d",
				"target": "https://example.com",
			},
		},
	}
	summary := FromSmartScan("https://example.com", scan)
	if len(summary.Findings) != 1 {
		t.Fatalf("findings len %d", len(summary.Findings))
	}
	if summary.Findings[0].Severity != domain.SeverityHigh {
		t.Fatalf("severity %s", summary.Findings[0].Severity)
	}
	br, _ := summary.Sections["severity_breakdown"].(map[string]int)
	if br["high"] != 1 {
		t.Fatalf("breakdown: %v", br)
	}
}
