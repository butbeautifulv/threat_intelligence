package report

import (
	"sort"
	"strings"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
)

// ExecutiveSummary is a structured assessment executive summary for agents.
type ExecutiveSummary struct {
	Target           string   `json:"target"`
	RiskPosture      string   `json:"risk_posture"`
	TotalFindings    int      `json:"total_findings"`
	Critical         int      `json:"critical"`
	High             int      `json:"high"`
	Medium           int      `json:"medium"`
	Low              int      `json:"low"`
	Info             int      `json:"info"`
	ToolsExecuted    int      `json:"tools_executed"`
	DurationSeconds  float64  `json:"duration_seconds"`
	TopRisks         []string `json:"top_risks"`
	Recommendations  []string `json:"recommendations"`
}

// BuildExecutiveSummary builds deterministic executive summary from scan output.
func BuildExecutiveSummary(target string, scan map[string]any, findings []domain.Finding, riskLevel string, technologies []string) ExecutiveSummary {
	br := SeverityBreakdown(findings)
	es := ExecutiveSummary{
		Target:          target,
		TotalFindings:   len(findings),
		Critical:        br["critical"],
		High:            br["high"],
		Medium:          br["medium"],
		Low:             br["low"],
		Info:            br["info"],
		RiskPosture:     riskPostureFromFindings(br, riskLevel),
		TopRisks:        topRiskTitles(findings, 5),
		Recommendations: recommendations(br, technologies),
	}
	if executed, ok := scan["tools_executed"].([]map[string]any); ok {
		es.ToolsExecuted = len(executed)
		var dur float64
		for _, e := range executed {
			if t, ok := e["execution_time"].(float64); ok {
				dur += t
			}
		}
		es.DurationSeconds = dur
	} else if executed, ok := scan["tools_executed"].([]any); ok {
		es.ToolsExecuted = len(executed)
	}
	return es
}

func riskPostureFromFindings(br map[string]int, profileRisk string) string {
	if br["critical"] > 0 {
		return "critical"
	}
	if br["high"] > 0 {
		return "high"
	}
	if br["medium"] > 0 {
		return "medium"
	}
	if profileRisk != "" {
		return strings.ToLower(profileRisk)
	}
	if br["low"] > 0 {
		return "low"
	}
	return "low"
}

func topRiskTitles(findings []domain.Finding, n int) []string {
	order := map[domain.Severity]int{
		domain.SeverityCritical: 0,
		domain.SeverityHigh:     1,
		domain.SeverityMedium:   2,
		domain.SeverityLow:      3,
	}
	sorted := append([]domain.Finding(nil), findings...)
	sort.Slice(sorted, func(i, j int) bool {
		oi, oj := order[sorted[i].Severity], order[sorted[j].Severity]
		if oi != oj {
			return oi < oj
		}
		return sorted[i].Title < sorted[j].Title
	})
	var out []string
	for _, f := range sorted {
		if f.Title == "" {
			continue
		}
		out = append(out, f.Title)
		if len(out) >= n {
			break
		}
	}
	return out
}

func recommendations(br map[string]int, technologies []string) []string {
	var rec []string
	if br["critical"] > 0 || br["high"] > 0 {
		rec = append(rec, "Prioritize remediation of critical/high findings before wider scanning")
		rec = append(rec, "Run nuclei/cve templates against confirmed vulnerable components")
	}
	if br["medium"] > 0 {
		rec = append(rec, "Schedule patching for medium-severity issues within normal change windows")
	}
	if len(technologies) > 0 {
		rec = append(rec, "Review technology-specific hardening guides for: "+strings.Join(technologies, ", "))
	}
	if len(rec) == 0 {
		rec = append(rec, "Continue monitoring; no high-severity findings in this assessment")
	}
	return rec
}
