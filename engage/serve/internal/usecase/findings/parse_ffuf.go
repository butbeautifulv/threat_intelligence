package findings

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

func parseFfuf(target, tool, output string) []domainreport.Finding {
	trim := strings.TrimSpace(output)
	if trim == "" {
		return nil
	}
	if trim[0] == '{' {
		var doc struct {
			Results []struct {
				URL    string            `json:"url"`
				Status int               `json:"status"`
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
