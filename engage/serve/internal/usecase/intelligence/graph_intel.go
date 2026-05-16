package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/hostnorm"
)

// TargetGraph returns unified veil-api graph state for agents and workflows.
func (s *Service) TargetGraph(ctx context.Context, target, indicators string) TargetGraphState {
	opts := TargetGraphLoadOpts{IncludeEngageContext: true}
	if q := strings.TrimSpace(indicators); q != "" {
		opts.SearchQuery = q
	}
	return LoadTargetGraph(ctx, s.Veil, target, opts)
}

// CorrelateThreatIntelligence merges target analysis with veil-graph search hits.
func (s *Service) CorrelateThreatIntelligence(ctx context.Context, target, indicators string) map[string]any {
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	out := map[string]any{
		"target":     target,
		"analysis":   analysis,
		"indicators": indicators,
		"success":    true,
	}
	query := hostnorm.NormalizeHost(target)
	if indicators != "" {
		query = strings.TrimSpace(indicators)
	}
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{
		IncludeEngageContext: true,
		SearchQuery:          query,
	})
	if state.GraphEnabled && query != "" {
		if len(state.Hits) > 0 {
			out["graph_hits"] = state.Hits
			out["correlation"] = "veil-graph"
		} else {
			out["correlation"] = "no_graph_hits"
		}
		if len(state.EngageContext) > 0 {
			out["engage_context"] = state.EngageContext
			out["related_cves"] = state.RelatedCVEs
		}
		if raw, ok := state.Hits["engage"]; ok {
			out["engage_findings"] = raw
		}
		if s.CVE != nil {
			details, merged := s.CVE.EnrichCorrelation(ctx, indicators, state.RelatedCVEs)
			if len(merged) > 0 {
				out["related_cves"] = merged
			}
			if len(details) > 0 {
				out["cve_intelligence"] = details
				out["cve_details"] = slimCVEDetails(details)
			}
		}
	} else {
		out["correlation"] = "heuristic_only"
		if s.CVE != nil {
			details, merged := s.CVE.EnrichCorrelation(ctx, indicators, nil)
			if len(merged) > 0 {
				out["related_cves"] = merged
			}
			if len(details) > 0 {
				out["cve_intelligence"] = details
				out["cve_details"] = slimCVEDetails(details)
			}
		}
	}
	out["graph_host"] = state.Host
	return out
}

func slimCVEDetails(details []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(details))
	for _, d := range details {
		row := map[string]any{}
		if id, ok := d["cve_id"].(string); ok {
			row["cve_id"] = id
		}
		if c, ok := d["cve"].(map[string]any); ok {
			if id, ok := c["cve_id"].(string); ok {
				row["cve_id"] = id
			}
			row["severity"] = c["severity"]
		}
		switch a := d["analysis"].(type) {
		case map[string]any:
			row["exploitability_score"] = a["exploitability_score"]
			row["exploitability_level"] = a["exploitability_level"]
			row["vulnerability_type"] = a["vulnerability_type"]
			if row["severity"] == nil {
				row["severity"] = a["severity"]
			}
		default:
			if b, err := json.Marshal(d["analysis"]); err == nil {
				var m map[string]any
				if json.Unmarshal(b, &m) == nil {
					row["exploitability_score"] = m["exploitability_score"]
					row["exploitability_level"] = m["exploitability_level"]
					row["vulnerability_type"] = m["vulnerability_type"]
				}
			}
		}
		if row["cve_id"] != nil {
			out = append(out, row)
		}
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
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{IncludeEngageContext: true})
	if state.GraphEnabled && state.Host != "" {
		if raw, ok := state.Hits["vuln"]; ok {
			out["graph_vuln_context"] = raw
		}
		if raw, ok := state.Hits["engage"]; ok {
			out["graph_engage_context"] = raw
		}
		if len(state.EngageContext) > 0 {
			out["engage_context"] = state.EngageContext
		}
	}
	cveIDs := append([]string{}, state.RelatedCVEs...)
	if raw, ok := state.Hits["vuln"]; ok {
		cveIDs = append(cveIDs, cve.ParseCVEIDs(string(raw))...)
	}
	cveIDs = append(cveIDs, cve.ParseCVEIDs(target)...)
	cveIDs = append(cveIDs, cve.ParseCVEIDs(objective)...)
	cveIDs = uniqueCVEIDs(cveIDs)
	if s.CVE != nil && len(cveIDs) > 0 {
		paths := s.CVE.BuildCVEAttackPaths(ctx, cveIDs)
		sortCVEPaths(paths)
		out["cve_attack_paths"] = paths
		out["cve_stages"] = cveStagesFromPaths(paths)
	}
	return out
}

func uniqueCVEIDs(ids []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, id := range ids {
		id = strings.ToUpper(strings.TrimSpace(id))
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func sortCVEPaths(paths []map[string]any) {
	sort.Slice(paths, func(i, j int) bool {
		return scoreFromPath(paths[i]) > scoreFromPath(paths[j])
	})
}

func scoreFromPath(p map[string]any) float64 {
	switch v := p["exploitability_score"].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		return 0
	}
}

func cveStagesFromPaths(paths []map[string]any) []map[string]any {
	stages := make([]map[string]any, 0, len(paths))
	for _, p := range paths {
		stages = append(stages, map[string]any{
			"cve_id":               p["cve_id"],
			"severity":             p["severity"],
			"exploitability_score": p["exploitability_score"],
			"exploit_available":    p["exploit_template_available"],
		})
	}
	return stages
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
		"target":          target,
		"analysis":        analysis,
		"tools_selected":  selected,
		"tools_executed":  len(results),
		"results":         results,
		"assessment_mode": "deterministic_ranked_scan",
		"note":            "not an LLM; uses DecisionEngine + catalog tools",
		"success":         true,
	}
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{IncludeEngageContext: true})
	if raw, ok := state.Hits["vuln"]; ok {
		out["graph_context"] = raw
	}
	if len(state.EngageContext) > 0 {
		out["engage_context"] = state.EngageContext
	}
	out["graph_host"] = state.Host
	var failed int
	for _, r := range results {
		if !r.Success {
			failed++
		}
	}
	out["summary"] = fmt.Sprintf("%d tools run, %d failed", len(results), failed)
	return out
}
