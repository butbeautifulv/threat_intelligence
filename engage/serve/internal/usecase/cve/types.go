package cve

// CVEEntry is a normalized CVE record from NVD or cache.
type CVEEntry struct {
	CVEID            string   `json:"cve_id"`
	Description      string   `json:"description"`
	Severity         string   `json:"severity"`
	CVSSScore        float64  `json:"cvss_score"`
	AttackVector     string   `json:"attack_vector,omitempty"`
	AttackComplexity string   `json:"attack_complexity,omitempty"`
	References       []string `json:"references,omitempty"`
	Published        string   `json:"published,omitempty"`
}

// ExploitabilityAnalysis is deterministic exploitability scoring for a CVE.
type ExploitabilityAnalysis struct {
	Success              bool    `json:"success"`
	CVEID                string  `json:"cve_id"`
	ExploitabilityScore  float64 `json:"exploitability_score"`
	ExploitabilityLevel  string  `json:"exploitability_level"`
	Severity             string  `json:"severity"`
	CVSSScore            float64 `json:"cvss_score"`
	AttackVector         string  `json:"attack_vector,omitempty"`
	VulnerabilityType    string  `json:"vulnerability_type,omitempty"`
	Recommendation       string  `json:"recommendation,omitempty"`
	Error                string  `json:"error,omitempty"`
}

// MonitorRequest configures CVE feed monitoring.
type MonitorRequest struct {
	Hours          int
	SeverityFilter string
	Keywords       string
	AnalyzeTop     int
}

// MonitorResult is the cve-monitor API response shape.
type MonitorResult struct {
	Success                bool                     `json:"success"`
	CVEMonitoring          map[string]any           `json:"cve_monitoring"`
	ExploitabilityAnalysis []ExploitabilityAnalysis `json:"exploitability_analysis"`
	Timestamp              string                   `json:"timestamp"`
	Error                  string                   `json:"error,omitempty"`
}

// ExploitRequest configures exploit template generation.
type ExploitRequest struct {
	CVEID         string
	Description   string
	TargetOS      string
	TargetArch    string
	ExploitType   string
	EvasionLevel  string
	TargetIP      string
	TargetPort    int
	Analysis      ExploitabilityAnalysis
}

// ExploitResult is the exploit-generate response shape.
type ExploitResult struct {
	Success           bool                   `json:"success"`
	CVEID             string                 `json:"cve_id,omitempty"`
	CVEAnalysis       ExploitabilityAnalysis `json:"cve_analysis,omitempty"`
	ExploitGeneration map[string]any         `json:"exploit_generation,omitempty"`
	ExistingExploits  []map[string]any       `json:"existing_exploits,omitempty"`
	Timestamp         string                 `json:"timestamp,omitempty"`
	Error             string                 `json:"error,omitempty"`
}
