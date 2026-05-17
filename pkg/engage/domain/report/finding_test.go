package report

import "testing"

func TestSeverityConstants(t *testing.T) {
	want := []Severity{
		SeverityInfo,
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}
	for _, s := range want {
		if s == "" {
			t.Fatal("severity constant must not be empty")
		}
	}
}
