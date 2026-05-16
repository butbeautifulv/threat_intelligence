package intelligence

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// TargetTimelineRequest configures aggregated target read-back.
type TargetTimelineRequest struct {
	Target       string
	Limit        int
	IncludeGraph bool
}

// TimelineEvent is one chronological entry in the merged timeline.
type TimelineEvent struct {
	At      string `json:"at"`
	Source  string `json:"source"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
}

// TargetTimelineResponse aggregates audit, graph, and correlation for a host.
type TargetTimelineResponse struct {
	Target         string         `json:"target"`
	Host           string         `json:"host"`
	Analysis       any            `json:"analysis,omitempty"`
	AuditEvents    []audit.Event  `json:"audit_events"`
	Graph          map[string]any `json:"graph,omitempty"`
	EngageContext  json.RawMessage `json:"engage_context,omitempty"`
	Correlation    map[string]any `json:"correlation,omitempty"`
	Timeline       []TimelineEvent `json:"timeline"`
	RelatedCVECount int           `json:"related_cve_count"`
}

// TargetTimeline builds a unified view for agents after scans.
func (s *Service) TargetTimeline(ctx context.Context, req TargetTimelineRequest) TargetTimelineResponse {
	target := strings.TrimSpace(req.Target)
	host := normalizeEngageHost(target)
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	resp := TargetTimelineResponse{
		Target:      target,
		Host:        host,
		AuditEvents: []audit.Event{},
		Graph:       map[string]any{},
		Timeline:    []TimelineEvent{},
	}
	if target == "" {
		return resp
	}
	resp.Analysis = s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})

	if s.Audit != nil {
		events, _ := s.Audit.Recent(limit * 2)
		for _, e := range events {
			if auditMatchesHost(e, host, target) {
				resp.AuditEvents = append(resp.AuditEvents, e)
				resp.Timeline = append(resp.Timeline, TimelineEvent{
					At:      e.At.UTC().Format(time.RFC3339),
					Source:  "audit",
					Kind:    "tool_run",
					Summary: e.Tool + " success=" + boolStr(e.Success),
				})
			}
		}
		if len(resp.AuditEvents) > limit {
			resp.AuditEvents = resp.AuditEvents[:limit]
		}
	}

	if s.Veil != nil && s.Veil.Enabled() && host != "" {
		for _, cat := range []string{"engage", "vuln", "ti"} {
			if raw, err := s.Veil.Search(ctx, cat, host); err == nil && len(raw) > 2 && string(raw) != "null" {
				resp.Graph[cat] = json.RawMessage(raw)
			}
		}
		if raw, err := s.Veil.EngageContext(ctx, host); err == nil && len(raw) > 2 {
			resp.EngageContext = raw
			resp.RelatedCVECount = countRelatedCVEs(raw)
			appendEngageContextTimeline(&resp, raw)
		}
	}
	if !req.IncludeGraph {
		resp.Graph = nil
	}
	resp.Correlation = s.CorrelateThreatIntelligence(ctx, target, "")
	sort.Slice(resp.Timeline, func(i, j int) bool {
		return resp.Timeline[i].At > resp.Timeline[j].At
	})
	if len(resp.Timeline) > limit {
		resp.Timeline = resp.Timeline[:limit]
	}
	return resp
}

func auditMatchesHost(e audit.Event, host, target string) bool {
	t := strings.ToLower(strings.TrimSpace(e.Target))
	if t == "" {
		return false
	}
	if host != "" && (strings.Contains(t, host) || t == host) {
		return true
	}
	return strings.Contains(t, strings.ToLower(target))
}

func normalizeEngageHost(target string) string {
	t := strings.TrimSpace(target)
	t = strings.TrimPrefix(t, "https://")
	t = strings.TrimPrefix(t, "http://")
	if i := strings.Index(t, "/"); i >= 0 {
		t = t[:i]
	}
	if i := strings.Index(t, ":"); i >= 0 {
		t = t[:i]
	}
	return strings.ToLower(strings.TrimSpace(t))
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func countRelatedCVEs(raw json.RawMessage) int {
	var wrap struct {
		Context struct {
			Vulnerabilities []any `json:"vulnerabilities"`
			Findings        []struct {
				RelatedVulnerabilities []any `json:"related_vulnerabilities"`
			} `json:"findings"`
		} `json:"context"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return 0
	}
	n := len(wrap.Context.Vulnerabilities)
	for _, f := range wrap.Context.Findings {
		n += len(f.RelatedVulnerabilities)
	}
	return n
}

func appendEngageContextTimeline(resp *TargetTimelineResponse, raw json.RawMessage) {
	var wrap struct {
		Context struct {
			ToolRuns []struct {
				Props map[string]any `json:"props"`
			} `json:"tool_runs"`
			Findings []struct {
				Node struct {
					Props map[string]any `json:"props"`
				} `json:"node"`
				RelatedVulnerabilities []struct {
					Props map[string]any `json:"props"`
				} `json:"related_vulnerabilities"`
			} `json:"findings"`
		} `json:"context"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return
	}
	for _, r := range wrap.Context.ToolRuns {
		at, tool := propsAtTool(r.Props)
		resp.Timeline = append(resp.Timeline, TimelineEvent{
			At: at, Source: "graph", Kind: "engage_tool_run", Summary: tool,
		})
	}
	for _, f := range wrap.Context.Findings {
		at, title := propsAtFinding(f.Node.Props)
		summary := title
		if len(f.RelatedVulnerabilities) > 0 {
			summary += " (+" + itoa(len(f.RelatedVulnerabilities)) + " CVE links)"
		}
		resp.Timeline = append(resp.Timeline, TimelineEvent{
			At: at, Source: "graph", Kind: "engage_finding", Summary: summary,
		})
	}
}

func propsAtTool(props map[string]any) (at, tool string) {
	if props == nil {
		return time.Now().UTC().Format(time.RFC3339), "tool"
	}
	if v, ok := props["at"].(string); ok {
		at = v
	}
	if v, ok := props["tool"].(string); ok {
		tool = v
	}
	if at == "" {
		at = time.Now().UTC().Format(time.RFC3339)
	}
	if tool == "" {
		tool = "tool_run"
	}
	return at, tool
}

func propsAtFinding(props map[string]any) (at, title string) {
	if props == nil {
		return time.Now().UTC().Format(time.RFC3339), "finding"
	}
	if v, ok := props["title"].(string); ok {
		title = v
	}
	if title == "" {
		title = "finding"
	}
	return time.Now().UTC().Format(time.RFC3339), title
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
