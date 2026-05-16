package report

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
