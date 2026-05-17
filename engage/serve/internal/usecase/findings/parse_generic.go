package findings

import (
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func parseGeneric(target, tool, output string) []domainreport.Finding {
	indicators := []struct {
		word     string
		severity domainreport.Severity
	}{
		{"CRITICAL", domainreport.SeverityCritical},
		{"HIGH", domainreport.SeverityHigh},
		{"MEDIUM", domainreport.SeverityMedium},
		{"VULNERABILITY", domainreport.SeverityHigh},
		{"SQL injection", domainreport.SeverityHigh},
		{"XSS", domainreport.SeverityMedium},
	}
	var out []domainreport.Finding
	low := strings.ToLower(output)
	for _, ind := range indicators {
		if strings.Contains(low, strings.ToLower(ind.word)) {
			out = append(out, domainreport.Finding{
				Title:       ind.word + " indicator",
				Severity:    ind.severity,
				Description: "matched output pattern",
				Target:      target,
				Tool:        tool,
			})
		}
	}
	return out
}
