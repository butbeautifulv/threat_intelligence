package findings

import (
	"regexp"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func parseSqlmap(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	low := strings.ToLower(output)
	if strings.Contains(low, "sqlmap identified the following injection point") {
		out = append(out, domainreport.Finding{
			Title:       "SQL injection point identified",
			Severity:    domainreport.SeverityHigh,
			Description: "sqlmap identified injectable parameter(s)",
			Target:      target,
			Tool:        tool,
		})
	}
	paramRe := regexp.MustCompile(`(?m)^Parameter:\s*(.+)$`)
	typeRe := regexp.MustCompile(`(?m)^\s+Type:\s*(.+)$`)
	titleRe := regexp.MustCompile(`(?m)^\s+Title:\s*(.+)$`)
	params := paramRe.FindAllStringSubmatch(output, -1)
	types := typeRe.FindAllStringSubmatch(output, -1)
	titles := titleRe.FindAllStringSubmatch(output, -1)
	for i, pm := range params {
		desc := strings.TrimSpace(pm[1])
		if i < len(types) {
			desc += " (" + strings.TrimSpace(types[i][1]) + ")"
		}
		title := "sqlmap: " + strings.TrimSpace(pm[1])
		if i < len(titles) && titles[i][1] != "" {
			title = "sqlmap: " + strings.TrimSpace(titles[i][1])
		}
		out = append(out, domainreport.Finding{
			Title:       title,
			Severity:    domainreport.SeverityHigh,
			Description: desc,
			Target:      target,
			Tool:        tool,
		})
	}
	if strings.Contains(low, "is vulnerable") {
		out = append(out, domainreport.Finding{
			Title:       "sqlmap: target is vulnerable",
			Severity:    domainreport.SeverityHigh,
			Description: "sqlmap reported vulnerability",
			Target:      target,
			Tool:        tool,
		})
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}
