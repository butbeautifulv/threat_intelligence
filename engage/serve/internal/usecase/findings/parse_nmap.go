package findings

import (
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func parseNmap(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "/tcp") && !strings.Contains(line, "/udp") {
			continue
		}
		if !strings.Contains(strings.ToLower(line), "open") {
			continue
		}
		out = append(out, domainreport.Finding{
			Title:       "open port",
			Severity:    domainreport.SeverityInfo,
			Description: strings.TrimSpace(line),
			Target:      target,
			Tool:        tool,
			Evidence:    line,
		})
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}
