package intelligence

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
	"github.com/butbeautifulv/veil/pkg/engage/hostnorm"
)

// DefaultGraphCategories are veil-api search categories for target decisions.
var DefaultGraphCategories = []string{"ti", "vuln", "engage"}

// TargetGraphLoadOpts configures LoadTargetGraph.
type TargetGraphLoadOpts struct {
	Categories           []string
	IncludeEngageContext bool
	SearchQuery          string // optional override for category search (e.g. CVE id)
}

// TargetGraphState is the unified graph read model for a target host.
type TargetGraphState struct {
	Target         string                     `json:"target"`
	Host           string                     `json:"host"`
	GraphEnabled   bool                       `json:"graph_enabled"`
	Hits           map[string]json.RawMessage `json:"hits,omitempty"`
	EngageContext  json.RawMessage            `json:"engage_context,omitempty"`
	EngageFound    bool                       `json:"engage_found"`
	RelatedCVEs    []string                   `json:"related_cves,omitempty"`
	RelatedCVECount int                       `json:"related_cve_count"`
}

// LoadTargetGraph loads category hits and optional engage subgraph via veil-api.
func LoadTargetGraph(ctx context.Context, veil veilgraph.Reader, target string, opts TargetGraphLoadOpts) TargetGraphState {
	state := TargetGraphState{
		Target: target,
		Hits:   map[string]json.RawMessage{},
	}
	state.Host = hostnorm.NormalizeHost(target)
	if veil == nil || !veil.Enabled() {
		return state
	}
	state.GraphEnabled = true
	query := opts.SearchQuery
	if query == "" {
		query = state.Host
	}
	if query == "" && !opts.IncludeEngageContext {
		return state
	}
	cats := opts.Categories
	if len(cats) == 0 {
		cats = DefaultGraphCategories
	}
	if query != "" {
		for _, cat := range cats {
			raw, err := veil.Search(ctx, cat, query)
			if err == nil && validGraphJSON(raw) {
				state.Hits[cat] = raw
			}
		}
	}
	if opts.IncludeEngageContext && state.Host != "" {
		raw, err := veil.EngageContext(ctx, state.Host)
		if err == nil && validGraphJSON(raw) {
			state.EngageContext = raw
			state.EngageFound = engageContextFound(raw)
			state.RelatedCVEs = ParseEngageContextCVEs(raw)
			state.RelatedCVECount = engageContextCVECount(raw)
		}
	}
	return state
}

func validGraphJSON(raw json.RawMessage) bool {
	return len(raw) > 2 && string(raw) != "null"
}

func engageContextFound(raw json.RawMessage) bool {
	var wrap struct {
		Found bool `json:"found"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return true
	}
	return wrap.Found
}

// ParseEngageContextCVEs extracts CVE ids linked in engage subgraph context.
func ParseEngageContextCVEs(raw json.RawMessage) []string {
	var wrap struct {
		Context struct {
			Vulnerabilities []struct {
				Props map[string]any `json:"props"`
			} `json:"vulnerabilities"`
		} `json:"context"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, v := range wrap.Context.Vulnerabilities {
		if v.Props == nil {
			continue
		}
		for _, key := range []string{"cve", "id"} {
			if c, ok := v.Props[key].(string); ok && c != "" {
				if _, dup := seen[c]; !dup {
					seen[c] = struct{}{}
					out = append(out, c)
				}
			}
		}
	}
	return out
}

func engageContextCVECount(raw json.RawMessage) int {
	var wrap struct {
		Context struct {
			Vulnerabilities []any `json:"vulnerabilities"`
			Findings        []struct {
				RelatedVulnerabilities []any `json:"related_vulnerabilities"`
			} `json:"findings"`
		} `json:"context"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return len(ParseEngageContextCVEs(raw))
	}
	n := len(wrap.Context.Vulnerabilities)
	for _, f := range wrap.Context.Findings {
		n += len(f.RelatedVulnerabilities)
	}
	return n
}
