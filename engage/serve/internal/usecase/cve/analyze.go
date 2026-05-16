package cve

import "strings"

// AnalyzeExploitability scores exploitability from CVE metadata (deterministic, no LLM).
func AnalyzeExploitability(entry CVEEntry) ExploitabilityAnalysis {
	a := ExploitabilityAnalysis{
		Success:     true,
		CVEID:       entry.CVEID,
		Severity:    entry.Severity,
		CVSSScore:   entry.CVSSScore,
		AttackVector: entry.AttackVector,
	}
	score := exploitabilityScore(entry)
	a.ExploitabilityScore = score
	a.ExploitabilityLevel = levelFromScore(score, entry.Severity)
	a.VulnerabilityType = ClassifyVuln(entry.Description)
	a.Recommendation = recommendation(a.ExploitabilityLevel, a.VulnerabilityType)
	return a
}

func exploitabilityScore(entry CVEEntry) float64 {
	if entry.CVSSScore > 0 {
		base := entry.CVSSScore / 10.0
		if base > 1 {
			base = 1
		}
		boost := keywordExploitBoost(entry.Description)
		score := base*0.7 + boost*0.3
		if score > 1 {
			return 1
		}
		return score
	}
	return keywordExploitBoost(entry.Description)
}

func keywordExploitBoost(desc string) float64 {
	low := strings.ToLower(desc)
	score := 0.2
	keywords := map[string]float64{
		"remote code execution": 0.35,
		"code execution":        0.3,
		"rce":                   0.3,
		"sql injection":         0.28,
		"command injection":     0.28,
		"authentication bypass": 0.25,
		"privilege escalation":  0.22,
		"buffer overflow":       0.22,
		"deserialization":       0.2,
		"xxe":                   0.18,
		"xss":                   0.12,
		"cross-site scripting":  0.12,
		"directory traversal":   0.15,
		"path traversal":        0.15,
	}
	for k, v := range keywords {
		if strings.Contains(low, k) {
			score += v
		}
	}
	if score > 1 {
		return 1
	}
	return score
}

func levelFromScore(score float64, severity string) string {
	switch {
	case score >= 0.85 || severity == "CRITICAL":
		return "CRITICAL"
	case score >= 0.65 || severity == "HIGH":
		return "HIGH"
	case score >= 0.4 || severity == "MEDIUM":
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func recommendation(level, vulnType string) string {
	switch level {
	case "CRITICAL", "HIGH":
		return "prioritize patching and validate with nuclei/cve templates; consider " + vulnType + " exploit PoC in lab only"
	case "MEDIUM":
		return "schedule remediation; run targeted scans for " + vulnType
	default:
		return "monitor; low immediate exploitability"
	}
}
