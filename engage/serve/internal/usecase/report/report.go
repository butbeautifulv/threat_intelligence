package report

import (
	"time"

	domain "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
)

// SummaryReport is a structured assessment summary for agents and APIs.
type SummaryReport struct {
	ReportType string         `json:"report_type"`
	Target     string         `json:"target"`
	Generated  time.Time      `json:"generated_at"`
	Sections   map[string]any `json:"sections"`
	Findings   []domain.Finding `json:"findings,omitempty"`
}

func NewSummary(target string, sections map[string]any, findings []domain.Finding) SummaryReport {
	return SummaryReport{
		ReportType: "summary",
		Target:     target,
		Generated:  time.Now().UTC(),
		Sections:   sections,
		Findings:   findings,
	}
}

// VulnerabilityCard is stable JSON for visual/vulnerability-card (no ANSI).
type VulnerabilityCard struct {
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Target      string `json:"target"`
	Tool        string `json:"tool,omitempty"`
	Evidence    string `json:"evidence,omitempty"`
}

func NewVulnerabilityCard(f domain.Finding) VulnerabilityCard {
	return VulnerabilityCard{
		Title:       f.Title,
		Severity:    string(f.Severity),
		Description: f.Description,
		Target:      f.Target,
		Tool:        f.Tool,
		Evidence:    f.Evidence,
	}
}

// ToolOutput wraps raw tool output for visual/tool-output.
type ToolOutput struct {
	Tool   string `json:"tool"`
	Target string `json:"target"`
	Output string `json:"output"`
	OK     bool   `json:"success"`
}
