package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func TestRenderAssessmentHTML_goldenFindingRow(t *testing.T) {
	summary := SummaryReport{
		ReportType: "summary",
		Target:     "https://example.com",
		Generated:  time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		Sections: map[string]any{
			"severity_breakdown": map[string]int{"high": 1},
		},
		Findings: []domain.Finding{
			{Title: "SQL injection risk", Severity: domain.SeverityHigh, Tool: "sqlmap_scan"},
		},
	}
	html := RenderAssessmentHTML(summary, DefaultBranding())
	wantPath := filepath.Join("testdata", "html_finding_row.fragment")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, strings.TrimSpace(string(want))) {
		t.Fatalf("html missing golden fragment from %s", wantPath)
	}
	if !strings.Contains(html, "Veil Engage") {
		t.Fatal("html missing branding")
	}
}

func TestRenderAssessmentHTML_goldenCustomBranding(t *testing.T) {
	summary := SummaryReport{
		ReportType: "summary",
		Target:     "https://lab.example",
		Generated:  time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
	}
	branding := Branding{
		Organization:   "Acme Red Team",
		Classification: "TLP:AMBER",
		LogoURL:        "https://example.com/logo.png",
		Footer:         "Acme internal use only",
	}
	html := RenderAssessmentHTML(summary, branding)
	wantPath := filepath.Join("testdata", "html_branding_logo.fragment")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, strings.TrimSpace(string(want))) {
		t.Fatalf("html missing golden fragment from %s", wantPath)
	}
	if !strings.Contains(html, "Acme Red Team") || !strings.Contains(html, "TLP:AMBER") {
		t.Fatal("html missing custom branding text")
	}
}
