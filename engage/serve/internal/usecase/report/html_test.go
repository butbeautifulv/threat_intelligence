package report

import (
	"strings"
	"testing"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func TestRenderAssessmentHTML_containsFinding(t *testing.T) {
	summary := NewSummary("https://example.com", map[string]any{
		"severity_breakdown": map[string]int{"high": 1},
	}, []domain.Finding{
		{Title: "SQL injection risk", Severity: domain.SeverityHigh, Tool: "sqlmap_scan"},
	})
	html := RenderAssessmentHTML(summary, DefaultBranding())
	if !strings.Contains(html, "SQL injection risk") {
		t.Fatal("html missing finding title")
	}
	if !strings.Contains(html, "Veil Engage") {
		t.Fatal("html missing branding")
	}
}
