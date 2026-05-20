package report

import (
	"testing"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func TestSeverityBreakdown(t *testing.T) {
	findings := []domain.Finding{
		{Severity: domain.SeverityCritical},
		{Severity: domain.SeverityHigh},
		{Severity: domain.SeverityMedium},
		{Severity: domain.SeverityLow},
		{Severity: domain.SeverityInfo},
		{Severity: domain.Severity("unrecognized")},
	}
	br := SeverityBreakdown(findings)
	want := map[string]int{
		"critical": 1,
		"high":     1,
		"medium":   1,
		"low":      1,
		"info":     2,
	}
	for k, v := range want {
		if br[k] != v {
			t.Fatalf("%s: got %d want %d; full map %v", k, br[k], v, br)
		}
	}
}
