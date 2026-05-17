package report

import (
	"testing"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func TestSeverityBreakdown(t *testing.T) {
	br := SeverityBreakdown([]domain.Finding{
		{Severity: domain.SeverityCritical},
		{Severity: domain.SeverityLow},
	})
	if br["critical"] != 1 || br["low"] != 1 {
		t.Fatalf("%v", br)
	}
}
