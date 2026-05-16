package report

import (
	"encoding/json"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
	findinguc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
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

// FromSmartScan builds a summary report from smart-scan output.
func FromSmartScan(target string, scan map[string]any) SummaryReport {
	sections := map[string]any{
		"scan_status": scan["status"],
		"objective":   scan["objective"],
	}
	if tools, ok := scan["tools_executed"].([]map[string]any); ok {
		sections["tools_executed"] = tools
	} else if tools, ok := scan["tools_executed"].([]any); ok {
		sections["tools_executed"] = tools
	}
	var rawFindings []domain.Finding
	if raw, ok := scan["findings"].([]domain.Finding); ok {
		rawFindings = raw
	} else if raw, ok := scan["findings"].([]any); ok {
		b, _ := json.Marshal(raw)
		_ = json.Unmarshal(b, &rawFindings)
	}
	rawFindings = findinguc.DedupeFindings(rawFindings)
	sections["severity_breakdown"] = SeverityBreakdown(rawFindings)
	summary := NewSummary(target, sections, rawFindings)
	return summary
}
