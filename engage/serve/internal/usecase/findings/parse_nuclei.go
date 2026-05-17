package findings

import (
	"encoding/json"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

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
