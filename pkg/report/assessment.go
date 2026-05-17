package report

import (
	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

// SeverityBreakdown counts findings by severity.
func SeverityBreakdown(findings []domain.Finding) map[string]int {
	out := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}
	for _, f := range findings {
		switch f.Severity {
		case domain.SeverityCritical:
			out["critical"]++
		case domain.SeverityHigh:
			out["high"]++
		case domain.SeverityMedium:
			out["medium"]++
		case domain.SeverityLow:
			out["low"]++
		default:
			out["info"]++
		}
	}
	return out
}
