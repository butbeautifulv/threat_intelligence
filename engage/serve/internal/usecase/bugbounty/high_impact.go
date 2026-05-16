package bugbounty

import (
	"sort"
	"strings"
)

// TestScenario is a named test with sample payloads (deterministic hints).
type TestScenario struct {
	Name     string   `json:"name"`
	Payloads []string `json:"payloads"`
}

// VulnProfile describes tools and scenarios for a vulnerability class.
type VulnProfile struct {
	Priority    int            `json:"priority"`
	Tools       []string       `json:"tools"`
	PayloadType string         `json:"payload_type"`
	Scenarios   []TestScenario `json:"test_scenarios,omitempty"`
}

// HighImpactVulns maps vuln types to hunting profiles (HexStrike L2451–2460).
var HighImpactVulns = map[string]VulnProfile{
	"rce":  {Priority: 10, Tools: []string{"nuclei", "sqlmap"}, PayloadType: "command_injection"},
	"sqli": {Priority: 9, Tools: []string{"sqlmap", "nuclei"}, PayloadType: "sql_injection"},
	"ssrf": {Priority: 8, Tools: []string{"nuclei", "ffuf"}, PayloadType: "ssrf"},
	"idor": {Priority: 8, Tools: []string{"arjun", "paramspider", "ffuf"}, PayloadType: "idor"},
	"xss":  {Priority: 7, Tools: []string{"dalfox", "nuclei"}, PayloadType: "xss"},
	"lfi":  {Priority: 7, Tools: []string{"ffuf", "nuclei"}, PayloadType: "lfi"},
	"xxe":  {Priority: 6, Tools: []string{"nuclei"}, PayloadType: "xxe"},
	"csrf": {Priority: 5, Tools: []string{"nuclei"}, PayloadType: "csrf"},
}

// testScenariosFor returns scenarios for a vulnerability type.
func testScenariosFor(vulnType string) []TestScenario {
	scenarios := map[string][]TestScenario{
		"rce": {
			{Name: "Command Injection", Payloads: []string{"$(whoami)", "`id`", ";ls -la"}},
		},
		"sqli": {
			{Name: "Union-based SQLi", Payloads: []string{"' UNION SELECT 1,2,3--", "' OR 1=1--"}},
		},
		"xss": {
			{Name: "Reflected XSS", Payloads: []string{"<script>alert(1)</script>"}},
		},
		"ssrf": {
			{Name: "Internal Network", Payloads: []string{"http://127.0.0.1:80"}},
		},
		"idor": {
			{Name: "Numeric IDOR", Payloads: []string{"id=1", "id=2"}},
		},
	}
	return scenarios[vulnType]
}

// SortedPriorityVulns returns vuln types sorted by high-impact priority descending.
func SortedPriorityVulns(types []string) []string {
	type pair struct {
		v string
		p int
	}
	var pairs []pair
	seen := map[string]struct{}{}
	for _, v := range types {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" {
			continue
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		p := 0
		if prof, ok := HighImpactVulns[v]; ok {
			p = prof.Priority
			pairs = append(pairs, pair{v, p})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].p > pairs[j].p
	})
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = p.v
	}
	return out
}
