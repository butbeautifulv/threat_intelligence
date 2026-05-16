package cve

import (
	"context"
	"regexp"
	"strings"
	"time"
)

var cveIDPattern = regexp.MustCompile(`CVE-\d{4}-\d{4,}`)

// Service is the CVE intelligence facade.
type Service struct {
	NVD  NVDClient
	Veil VeilSearcher
}

// NewService wires CVE dependencies.
func NewService(veil VeilSearcher, nvd NVDClient) *Service {
	if nvd == nil {
		nvd = DefaultNVDClient()
	}
	return &Service{NVD: nvd, Veil: veil}
}

// GenerateExploitFromCVE looks up CVE, analyzes, and returns exploit templates.
func (s *Service) GenerateExploitFromCVE(ctx context.Context, body map[string]any) ExploitResult {
	now := time.Now().UTC().Format(time.RFC3339)
	cveID := normalizeCVEID(strVal(body, "cve_id"))
	if cveID == "" {
		return ExploitResult{Success: false, Error: "cve_id required", Timestamp: now}
	}
	if s.NVD == nil {
		return ExploitResult{Success: false, Error: "NVD client not configured", Timestamp: now}
	}
	entry, err := s.NVD.FetchCVE(ctx, cveID)
	if err != nil {
		return ExploitResult{Success: false, Error: err.Error(), CVEID: cveID, Timestamp: now}
	}
	analysis := AnalyzeExploitability(*entry)
	req := ExploitRequest{
		CVEID:        cveID,
		Description:  entry.Description,
		TargetOS:     strVal(body, "target_os"),
		TargetArch:   strVal(body, "target_arch", "x64"),
		ExploitType:  strVal(body, "exploit_type", "poc"),
		EvasionLevel: strVal(body, "evasion_level", "none"),
		TargetIP:     strVal(body, "target_ip", "127.0.0.1"),
		TargetPort:   intVal(body, "target_port", 80),
		Analysis:     analysis,
	}
	out := GenerateExploit(req)
	out.Timestamp = now
	out.ExistingExploits = []map[string]any{
		{"source": "template", "note": "deterministic PoC; verify in lab", "reliability": "UNVERIFIED"},
	}
	return out
}

// MonitorFromBody parses HTTP/MCP body for monitor feeds.
func (s *Service) MonitorFromBody(ctx context.Context, body map[string]any) MonitorResult {
	return s.MonitorFeeds(ctx, MonitorRequest{
		Hours:          intVal(body, "hours", 24),
		SeverityFilter: strVal(body, "severity_filter", "HIGH,CRITICAL"),
		Keywords:       strVal(body, "keywords", ""),
		AnalyzeTop:     intVal(body, "analyze_top", 5),
	})
}

// ParseCVEIDs extracts CVE identifiers from a string.
func ParseCVEIDs(s string) []string {
	found := cveIDPattern.FindAllString(strings.ToUpper(s), -1)
	seen := map[string]struct{}{}
	var out []string
	for _, id := range found {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// LookupSummary returns minimal detail for correlate (avoids full map when NVD fails).
func (s *Service) LookupSummary(ctx context.Context, cveID string) map[string]any {
	if s == nil {
		return nil
	}
	out := s.Lookup(ctx, cveID)
	if out["success"] != true {
		return map[string]any{
			"cve_id": cveID,
			"error":  strVal(out, "error"),
		}
	}
	return out
}

// BuildCVEAttackPaths builds ordered paths for discover-attack-chains.
func (s *Service) BuildCVEAttackPaths(ctx context.Context, cveIDs []string) []map[string]any {
	if s == nil || len(cveIDs) == 0 {
		return nil
	}
	var paths []map[string]any
	for _, id := range cveIDs {
		level := "UNKNOWN"
		score := 0.0
		vulnType := "generic"
		available := false
		if s.NVD != nil {
			if entry, err := s.NVD.FetchCVE(ctx, id); err == nil {
				a := AnalyzeExploitability(*entry)
				level = a.ExploitabilityLevel
				score = a.ExploitabilityScore
				vulnType = a.VulnerabilityType
				available = true
			}
		}
		paths = append(paths, map[string]any{
			"cve_id":                     id,
			"severity":                   level,
			"exploitability_score":       score,
			"vulnerability_type":         vulnType,
			"suggested_tools":            suggestedTools(vulnType),
			"exploit_template_available": available,
		})
	}
	return paths
}

func suggestedTools(vulnType string) []string {
	switch vulnType {
	case "sql_injection":
		return []string{"sqlmap", "nuclei"}
	case "xss":
		return []string{"dalfox", "nuclei"}
	case "rce":
		return []string{"nuclei", "nmap"}
	default:
		return []string{"nuclei", "searchsploit"}
	}
}

func strVal(m map[string]any, k string, def ...string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func intVal(m map[string]any, k string, def int) int {
	switch v := m[k].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return def
	}
}
