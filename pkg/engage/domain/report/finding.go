package report

import (
	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
)

// Severity classifies a finding.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Finding is a single security finding.
type Finding struct {
	Title       string   `json:"title"`
	Severity    Severity `json:"severity"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Tool        string   `json:"tool,omitempty"`
	Evidence    string   `json:"evidence,omitempty"`
}

// ToFindingEvent converts a finding to the NATS wire event shape.
func (f Finding) ToFindingEvent() engageevents.FindingEvent {
	return engageevents.FindingEvent{
		Tool:        f.Tool,
		Target:      f.Target,
		Title:       f.Title,
		Severity:    string(f.Severity),
		Description: f.Description,
	}
}
