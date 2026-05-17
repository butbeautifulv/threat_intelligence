package findings

import (
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

// ParseToolOutput extracts structured findings from tool stdout.
func ParseToolOutput(toolName, target, output string) []domainreport.Finding {
	low := strings.ToLower(toolName)
	var parsed []domainreport.Finding
	switch {
	case strings.Contains(low, "nuclei"):
		parsed = parseNuclei(target, toolName, output)
	case strings.Contains(low, "nmap"):
		parsed = parseNmap(target, toolName, output)
	case strings.Contains(low, "masscan"):
		parsed = parseMasscan(target, toolName, output)
	case strings.Contains(low, "ffuf"):
		parsed = parseFfuf(target, toolName, output)
	case strings.Contains(low, "sqlmap"):
		parsed = parseSqlmap(target, toolName, output)
	case strings.Contains(low, "wpscan"):
		parsed = parseWpscan(target, toolName, output)
	default:
		parsed = parseGeneric(target, toolName, output)
	}
	return DedupeFindings(parsed)
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
