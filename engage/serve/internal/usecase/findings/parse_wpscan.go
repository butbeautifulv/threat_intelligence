package findings

import (
	"encoding/json"
	"regexp"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

var wpscanTitleLine = regexp.MustCompile(`^\s*\[!]\s+(.*)$`)

func parseWpscan(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	trim := strings.TrimSpace(output)
	if trim != "" && trim[0] == '{' {
		type vulnStub struct {
			Title       string `json:"title"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		var doc struct {
			Vulnerabilities []vulnStub `json:"vulnerabilities"`
			Interesting     []struct {
				URL   string `json:"url"`
				Found string `json:"found_by"`
			} `json:"interesting_findings"`
		}
		if err := json.Unmarshal([]byte(trim), &doc); err == nil {
			for _, v := range doc.Vulnerabilities {
				title := strings.TrimSpace(v.Title)
				if title == "" {
					title = strings.TrimSpace(v.Name)
				}
				if title == "" {
					continue
				}
				desc := strings.TrimSpace(v.Description)
				out = append(out, domainreport.Finding{
					Title:       "wpscan: " + title,
					Severity:    domainreport.SeverityMedium,
					Description: desc,
					Target:      target,
					Tool:        tool,
				})
			}
			for _, it := range doc.Interesting {
				if it.URL == "" {
					continue
				}
				out = append(out, domainreport.Finding{
					Title:       "wpscan: " + strings.TrimSpace(it.URL),
					Severity:    domainreport.SeverityInfo,
					Description: strings.TrimSpace(it.Found),
					Target:      target,
					Tool:        tool,
					Evidence:    strings.TrimSpace(it.URL),
				})
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if m := wpscanTitleLine.FindStringSubmatch(line); len(m) == 2 {
			msg := strings.TrimSpace(m[1])
			if msg == "" {
				continue
			}
			low := strings.ToLower(msg)
			sev := domainreport.SeverityLow
			if strings.Contains(low, "critical") {
				sev = domainreport.SeverityCritical
			} else if strings.Contains(low, "high") {
				sev = domainreport.SeverityHigh
			} else if strings.Contains(low, "medium") {
				sev = domainreport.SeverityMedium
			}
			out = append(out, domainreport.Finding{
				Title:       "wpscan: " + msg,
				Severity:    sev,
				Description: msg,
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
