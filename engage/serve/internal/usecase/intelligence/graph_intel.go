package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// CorrelateThreatIntelligence merges target analysis with veil-graph search hits.
func (s *Service) CorrelateThreatIntelligence(ctx context.Context, target, indicators string) map[string]any {
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	out := map[string]any{
		"target":     target,
		"analysis":   analysis,
		"indicators": indicators,
		"success":    true,
	}
	query := graphSearchQuery(target)
	if indicators != "" {
		query = strings.TrimSpace(indicators)
	}
	if s.Veil != nil && s.Veil.Enabled() && query != "" {
		hits := map[string]json.RawMessage{}
		for _, cat := range []string{"ti", "vuln", "ioc"} {
			if raw, err := s.Veil.Search(ctx, cat, query); err == nil && len(raw) > 2 && string(raw) != "null" {
				hits[cat] = raw
			}
		}
		if len(hits) > 0 {
			out["graph_hits"] = hits
			out["correlation"] = "veil-graph"
		} else {
			out["correlation"] = "no_graph_hits"
		}
	} else {
		out["correlation"] = "heuristic_only"
	}
	return out
}

// DiscoverAttackChains returns analysis, attack chain plan, and optional graph context.
func (s *Service) DiscoverAttackChains(ctx context.Context, target, objective string) map[string]any {
	if objective == "" {
		objective = "comprehensive"
	}
	chain := s.CreateAttackChain(ctx, target, objective)
	out := map[string]any{
		"target":       target,
		"objective":    objective,
		"attack_chain": chain,
		"success":      true,
	}
	if s.Veil != nil && s.Veil.Enabled() {
		host := graphSearchQuery(target)
		if host != "" {
			if raw, err := s.Veil.Search(ctx, "vuln", host); err == nil && len(raw) > 2 {
				out["graph_vuln_context"] = raw
			}
		}
	}
	return out
}

// AIVulnerabilityAssessment runs a deterministic ranked scan (not LLM).
func (s *Service) AIVulnerabilityAssessment(ctx context.Context, subject, target string, maxTools int) map[string]any {
	if maxTools <= 0 {
		maxTools = 6
	}
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	selected := s.SelectToolsForTarget(ctx, analysis.TargetType, "comprehensive", target)
	if len(selected) > maxTools {
		selected = selected[:maxTools]
	}
	results := make([]contract.ToolRunResponse, 0, len(selected))
	if s.Tools != nil {
		for _, toolName := range selected {
			params := s.OptimizeParameters(analysis.TargetType, toolName, map[string]string{"target": target})
			res := s.Tools.Run(ctx, subject, toolName, contract.ToolRunRequest{
				Target:     target,
				Parameters: params,
			})
			results = append(results, res)
		}
	}
	out := map[string]any{
		"target":            target,
		"analysis":          analysis,
		"tools_selected":    selected,
		"tools_executed":    len(results),
		"results":           results,
		"assessment_mode":   "deterministic_ranked_scan",
		"note":              "not an LLM; uses DecisionEngine + catalog tools",
		"success":           true,
	}
	if s.Veil != nil && s.Veil.Enabled() {
		host := graphSearchQuery(target)
		if host != "" {
			if raw, err := s.Veil.Search(ctx, "vuln", host); err == nil {
				out["graph_context"] = json.RawMessage(raw)
			}
		}
	}
	var failed int
	for _, r := range results {
		if !r.Success {
			failed++
		}
	}
	out["summary"] = fmt.Sprintf("%d tools run, %d failed", len(results), failed)
	return out
}
