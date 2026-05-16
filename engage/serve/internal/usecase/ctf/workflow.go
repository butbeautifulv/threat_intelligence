package ctf

import (
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
)

// WorkflowStep is one step in a CTF challenge workflow.
type WorkflowStep struct {
	Step          int      `json:"step"`
	Action        string   `json:"action"`
	Description   string   `json:"description"`
	Parallel      bool     `json:"parallel"`
	Tools         []string `json:"tools"`
	EstimatedTime int      `json:"estimated_time"`
}

// ChallengeWorkflow is the structured plan for solving a challenge.
type ChallengeWorkflow struct {
	Challenge           string         `json:"challenge"`
	Category            string         `json:"category"`
	Difficulty          string         `json:"difficulty"`
	Points              int            `json:"points"`
	Tools               []string       `json:"tools"`
	Strategies          []Strategy     `json:"strategies"`
	WorkflowSteps       []WorkflowStep `json:"workflow_steps"`
	EstimatedTime       int            `json:"estimated_time"`
	SuccessProbability  float64        `json:"success_probability"`
	AutomationLevel     string         `json:"automation_level"`
	ParallelTasks       []string       `json:"parallel_tasks"`
	ValidationSteps     []string       `json:"validation_steps"`
}

// Strategy is a category-specific solving approach.
type Strategy struct {
	Strategy    string `json:"strategy"`
	Description string `json:"description"`
}

var solvingStrategies = map[string][]Strategy{
	"web": {
		{Strategy: "source_code_analysis", Description: "Analyze HTML/JS source for hidden information"},
		{Strategy: "sql_injection", Description: "Test for SQL injection in parameters"},
		{Strategy: "authentication_bypass", Description: "Test for auth bypass techniques"},
	},
	"crypto": {
		{Strategy: "frequency_analysis", Description: "Frequency analysis for substitution ciphers"},
		{Strategy: "weak_keys", Description: "Test for weak cryptographic keys"},
	},
	"pwn": {
		{Strategy: "buffer_overflow", Description: "Exploit buffer overflow vulnerabilities"},
		{Strategy: "rop_chains", Description: "Build ROP chains for exploitation"},
	},
	"forensics": {
		{Strategy: "steganography", Description: "Extract hidden data from images/audio"},
		{Strategy: "metadata_analysis", Description: "Analyze file metadata"},
	},
	"rev": {
		{Strategy: "static_analysis", Description: "Analyze binary without execution"},
		{Strategy: "dynamic_analysis", Description: "Analyze binary during execution"},
	},
}

// Manager builds CTF workflows.
type Manager struct {
	Tools *ToolManager
}

func NewManager() *Manager {
	return &Manager{Tools: NewToolManager()}
}

// CreateChallengeWorkflow builds a workflow for a challenge.
func (m *Manager) CreateChallengeWorkflow(ch Challenge, resolvedTools []string) ChallengeWorkflow {
	if len(resolvedTools) == 0 {
		resolvedTools = m.Tools.SuggestTools(ch.Description, ch.Category)
	}
	wf := ChallengeWorkflow{
		Challenge:          ch.Name,
		Category:           ch.Category,
		Difficulty:         ch.Difficulty,
		Points:             ch.Points,
		Tools:              resolvedTools,
		Strategies:         solvingStrategies[ch.Category],
		WorkflowSteps:      categoryWorkflowSteps(ch.Category),
		EstimatedTime:      estimateTime(ch),
		SuccessProbability: estimateSuccess(ch, len(resolvedTools)),
		AutomationLevel:    "high",
		ParallelTasks:      parallelTasksFor(ch.Category),
		ValidationSteps:    []string{"verify_flag_format", "submit_flag"},
	}
	return wf
}

func categoryWorkflowSteps(category string) []WorkflowStep {
	switch NormalizeCategory(category) {
	case "web":
		return []WorkflowStep{
			{Step: 1, Action: "automated_reconnaissance", Description: "Web recon and tech detect", Parallel: true, Tools: []string{"httpx", "katana"}, EstimatedTime: 300},
			{Step: 2, Action: "directory_enumeration", Description: "Directory enumeration", Parallel: true, Tools: []string{"gobuster", "feroxbuster"}, EstimatedTime: 900},
			{Step: 3, Action: "vulnerability_scanning", Description: "Vuln scan", Parallel: true, Tools: []string{"nuclei", "sqlmap", "dalfox"}, EstimatedTime: 1200},
			{Step: 4, Action: "flag_extraction", Description: "Extract flag", Parallel: false, Tools: []string{"manual"}, EstimatedTime: 300},
		}
	case "pwn":
		return []WorkflowStep{
			{Step: 1, Action: "binary_reconnaissance", Description: "Binary protections and strings", Parallel: true, Tools: []string{"checksec", "strings", "file"}, EstimatedTime: 600},
			{Step: 2, Action: "static_analysis", Description: "Static disassembly", Parallel: true, Tools: []string{"ghidra", "radare2"}, EstimatedTime: 1800},
			{Step: 3, Action: "exploit_development", Description: "Develop exploit", Parallel: false, Tools: []string{"pwntools"}, EstimatedTime: 2400},
		}
	case "crypto":
		return []WorkflowStep{
			{Step: 1, Action: "cipher_identification", Description: "Identify cipher type", Parallel: false, Tools: []string{"hash-identifier"}, EstimatedTime: 300},
			{Step: 2, Action: "automated_attacks", Description: "Automated crypto attacks", Parallel: true, Tools: []string{"hashcat", "john"}, EstimatedTime: 1800},
		}
	case "forensics":
		return []WorkflowStep{
			{Step: 1, Action: "file_analysis", Description: "File structure analysis", Parallel: true, Tools: []string{"binwalk", "strings", "exiftool"}, EstimatedTime: 900},
			{Step: 2, Action: "steganography_detection", Description: "Stego detection", Parallel: true, Tools: []string{"steghide", "zsteg"}, EstimatedTime: 1200},
		}
	default:
		return []WorkflowStep{
			{Step: 1, Action: "challenge_analysis", Description: "Analyze challenge requirements", Parallel: false, Tools: []string{"manual"}, EstimatedTime: 300},
			{Step: 2, Action: "solution_implementation", Description: "Implement solution", Parallel: false, Tools: []string{"manual"}, EstimatedTime: 900},
		}
	}
}

func estimateTime(ch Challenge) int {
	base := map[string]int{"easy": 1800, "medium": 3600, "hard": 7200, "insane": 14400, "unknown": 5400}
	sec := base[strings.ToLower(ch.Difficulty)]
	mult := map[string]float64{"web": 1.0, "crypto": 1.3, "pwn": 1.5, "forensics": 1.2, "rev": 1.4, "misc": 0.8, "osint": 0.9}
	if m, ok := mult[ch.Category]; ok {
		sec = int(float64(sec) * m)
	}
	return sec
}

func estimateSuccess(ch Challenge, toolCount int) float64 {
	base := map[string]float64{"easy": 0.85, "medium": 0.65, "hard": 0.45, "insane": 0.25, "unknown": 0.55}
	p := base[strings.ToLower(ch.Difficulty)]
	p += float64(toolCount) * 0.02
	if p > 0.95 {
		p = 0.95
	}
	return p
}

func parallelTasksFor(category string) []string {
	switch NormalizeCategory(category) {
	case "web":
		return []string{"recon", "dir_enum", "param_discovery"}
	case "pwn":
		return []string{"checksec", "strings"}
	default:
		return nil
	}
}

// MergePatternTools adds tools from intelligence attack patterns when target is set.
func MergePatternTools(ch Challenge, suggested []string) []string {
	if ch.TargetOrURL() == "" {
		return suggested
	}
	key := intelligence.SelectPatternKey("web", ch.Category)
	if ch.Category == "pwn" || ch.Category == "rev" {
		key = intelligence.SelectPatternKey("binary", "pwn")
	}
	patterns := intelligence.AttackPatterns()
	steps, ok := patterns[key]
	if !ok {
		return suggested
	}
	seen := map[string]struct{}{}
	for _, t := range suggested {
		seen[t] = struct{}{}
	}
	var out []string
	out = append(out, suggested...)
	for _, step := range steps {
		if _, dup := seen[step.Tool]; dup {
			continue
		}
		seen[step.Tool] = struct{}{}
		out = append(out, step.Tool)
	}
	return out
}
