package bugbounty

import "strings"

// Target describes a bug bounty program target (HexStrike BugBountyTarget parity).
type Target struct {
	Domain         string   `json:"domain"`
	Scope          []string `json:"scope,omitempty"`
	OutOfScope     []string `json:"out_of_scope,omitempty"`
	ProgramType    string   `json:"program_type"`
	PriorityVulns  []string `json:"priority_vulns,omitempty"`
	BountyRange    string   `json:"bounty_range,omitempty"`
	IncludeOSINT   bool     `json:"include_osint,omitempty"`
	IncludeBusiness bool    `json:"include_business_logic,omitempty"`
}

// RunOptions configures plan vs execute behavior.
type RunOptions struct {
	Execute bool
	Async   bool
	MaxTools int
}

// TargetFromBody parses target fields from HTTP/MCP JSON (domain, target, or target_url).
func TargetFromBody(body map[string]any) Target {
	t := Target{
		Domain:        firstString(body, "domain", "target", "target_url"),
		ProgramType:   str(body, "program_type"),
		BountyRange:   str(body, "bounty_range"),
		IncludeOSINT:  boolVal(body, "include_osint", true),
		IncludeBusiness: boolVal(body, "include_business_logic", true),
	}
	if t.ProgramType == "" {
		t.ProgramType = "web"
	}
	if pv, ok := body["priority_vulns"].([]any); ok {
		for _, v := range pv {
			if s, ok := v.(string); ok {
				t.PriorityVulns = append(t.PriorityVulns, s)
			}
		}
	}
	if len(t.PriorityVulns) == 0 {
		t.PriorityVulns = []string{"rce", "sqli", "xss", "idor", "ssrf"}
	}
	if scope, ok := body["scope"].([]any); ok {
		for _, v := range scope {
			if s, ok := v.(string); ok {
				t.Scope = append(t.Scope, s)
			}
		}
	}
	return t
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v := str(m, k); v != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func str(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func boolVal(m map[string]any, k string, def bool) bool {
	if v, ok := m[k].(bool); ok {
		return v
	}
	return def
}
