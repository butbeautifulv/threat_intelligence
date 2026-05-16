package findings

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
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
	case strings.Contains(low, "ffuf"):
		return parseFfuf(target, toolName, output)
	case strings.Contains(low, "sqlmap"):
		return parseSqlmap(target, toolName, output)
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

func parseFfuf(target, tool, output string) []domainreport.Finding {
	trim := strings.TrimSpace(output)
	if trim == "" {
		return nil
	}
	if trim[0] == '{' {
		var doc struct {
			Results []struct {
				URL    string `json:"url"`
				Status int    `json:"status"`
				Input  map[string]string `json:"input"`
			} `json:"results"`
		}
		if err := json.Unmarshal([]byte(trim), &doc); err == nil && len(doc.Results) > 0 {
			var out []domainreport.Finding
			for _, r := range doc.Results {
				title := r.URL
				if title == "" {
					for _, v := range r.Input {
						title = v
						break
					}
				}
				if title == "" {
					continue
				}
				out = append(out, domainreport.Finding{
					Title:       "ffuf: " + title,
					Severity:    ffufSeverity(r.Status),
					Description: fmt.Sprintf("status %d", r.Status),
					Target:      target,
					Tool:        tool,
					Evidence:    fmt.Sprintf(`{"url":%q,"status":%d}`, title, r.Status),
				})
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	var out []domainreport.Finding
	statusRe := regexp.MustCompile(`(?i)status:\s*(\d+)`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := statusRe.FindStringSubmatch(line)
		if len(m) == 2 {
			code, _ := strconv.Atoi(m[1])
			out = append(out, domainreport.Finding{
				Title:       "ffuf match",
				Severity:    ffufSeverity(code),
				Description: line,
				Target:      target,
				Tool:        tool,
				Evidence:    line,
			})
		}
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func ffufSeverity(status int) domainreport.Severity {
	switch {
	case status >= 500:
		return domainreport.SeverityHigh
	case status == 401 || status == 403:
		return domainreport.SeverityMedium
	case status >= 200 && status < 300:
		return domainreport.SeverityLow
	default:
		return domainreport.SeverityInfo
	}
}

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
