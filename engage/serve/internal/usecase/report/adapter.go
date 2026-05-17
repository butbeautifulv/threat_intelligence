package report

import (
	"encoding/json"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	findinguc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	pkgreport "github.com/butbeautifulv/veil/pkg/report"
)

// FromSmartScan builds a summary report from smart-scan output (engage adapter).
func FromSmartScan(target string, scan map[string]any) pkgreport.SummaryReport {
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
	sections["severity_breakdown"] = pkgreport.SeverityBreakdown(rawFindings)
	return pkgreport.NewSummary(target, sections, rawFindings)
}
