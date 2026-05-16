package findings

import (
	"encoding/json"
	"strings"

	domainreport "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
)

// ParseToolOutput extracts structured findings from tool stdout.
func ParseToolOutput(toolName, target, output string) []domainreport.Finding {
	low := strings.ToLower(toolName)
	switch {
	case strings.Contains(low, "nuclei"):
		return parseNuclei(target, toolName, output)
	case strings.Contains(low, "nmap"):
		return parseNmap(target, toolName, output)
	default:
		return parseGeneric(target, toolName, output)
	}
}

func parseNuclei(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var rec struct {
			Info struct {
				Name     string `json:"name"`
				Severity string `json:"severity"`
			} `json:"info"`
			MatcherName string `json:"matcher-name"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		title := rec.Info.Name
		if title == "" {
			title = rec.MatcherName
		}
		if title == "" {
			continue
		}
		out = append(out, domainreport.Finding{
			Title:       title,
			Severity:    mapSeverity(rec.Info.Severity),
			Description: title,
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

func mapSeverity(s string) domainreport.Severity {
	switch strings.ToLower(s) {
	case "critical":
		return domainreport.SeverityCritical
	case "high":
		return domainreport.SeverityHigh
	case "medium":
		return domainreport.SeverityMedium
	case "low":
		return domainreport.SeverityLow
	default:
		return domainreport.SeverityInfo
	}
}

// Count returns total findings across tool results.
func Count(all []domainreport.Finding) int {
	return len(all)
}
