package intelligence

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ParallelToolRunner runs multiple catalog tools concurrently (workflow.Service).
type ParallelToolRunner interface {
	RunToolsParallel(ctx context.Context, subject, target, targetType string, toolNames []string) []map[string]any
}

// Service provides target analysis and tool selection.
type Service struct {
	Veil            veilgraph.Reader
	Registry        *tools.Registry
	Engine          *DecisionEngine
	Tools           *toolsuc.Runner
	Audit           audit.Reader
	CVE             CVEIntelligence
	ParallelRunner  ParallelToolRunner
}

// CVEIntelligence is implemented by cve.Service (defined here to avoid import cycles in tests).
type CVEIntelligence interface {
	EnrichCorrelation(ctx context.Context, indicators string, related []string) ([]map[string]any, []string)
	BuildCVEAttackPaths(ctx context.Context, cveIDs []string) []map[string]any
}

func (s *Service) engine() *DecisionEngine {
	if s.Engine != nil {
		return s.Engine
	}
	return DefaultDecisionEngine()
}

func (s *Service) AnalyzeTarget(ctx context.Context, req contract.AnalyzeTargetRequest) contract.AnalyzeTargetResponse {
	target := strings.TrimSpace(req.Target)
	tt, tech, cms, _, probeHdr, probeBody := probeTarget(ctx, target)
	ips := resolveTargetIPs(ctx, target)
	techLabels := technologiesDetected(tech, cms)
	profile := BuildTargetProfile(target, tt, techLabels, cms, ips, 0)
	resp := contract.AnalyzeTargetResponse{
		Target:       target,
		TargetType:   tt,
		Technologies: tech,
		RiskLevel:    profile.RiskLevel,
		Confidence:   profile.ConfidenceScore,
		Metadata:     map[string]any{},
	}
	if cms != "" {
		resp.Metadata["cms"] = cms
	}
	resp.Metadata["attack_surface_score"] = profile.AttackSurfaceScore
	resp.Metadata["ip_addresses"] = profile.IPAddresses
	stack := DetectTechnologies(ctx, target, probeHdr, probeBody)
	resp.Metadata["technologies_detected"] = TechnologiesToStrings(stack)
	resp.Metadata["technology_stack"] = TechnologiesToStrings(stack)
	if s.Veil != nil && s.Veil.Enabled() {
		if raw, err := s.Veil.Categories(ctx); err == nil {
			resp.Metadata["veil_categories"] = raw
			if resp.Confidence < 0.85 {
				resp.Confidence = 0.85
			}
		}
		s.enrichGraph(ctx, target, resp.Metadata)
		if boost := s.graphBoost(ctx, target); len(boost) > 0 {
			resp.Metadata["graph_tool_boost"] = boost
		}
	}
	return resp
}

// BuildProfileFromTarget runs probe + profile assembly for internal engine use.
func (s *Service) BuildProfileFromTarget(ctx context.Context, target string) TargetProfile {
	tt, tech, cms, _, _, _ := probeTarget(ctx, target)
	ips := resolveTargetIPs(ctx, target)
	labels := technologiesDetected(tech, cms)
	return BuildTargetProfile(target, tt, labels, cms, ips, 0)
}

func (s *Service) enrichGraph(ctx context.Context, target string, meta map[string]any) {
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{})
	if len(state.Hits) > 0 {
		meta["graph_hits"] = state.Hits
		meta["graph_vuln_context"] = true
		meta["graph_host"] = state.Host
	}
}

func (s *Service) graphBoost(ctx context.Context, target string) map[string]float64 {
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{})
	if !state.GraphEnabled || state.Host == "" {
		return nil
	}
	boost := map[string]float64{}
	for cat, tools := range map[string][]string{
		"vuln":   {"nuclei", "nikto", "sqlmap"},
		"ti":     {"nuclei", "httpx"},
		"engage": {"nuclei", "nmap", "httpx"},
	} {
		if _, ok := state.Hits[cat]; !ok {
			continue
		}
		for _, t := range tools {
			boost[t] += 0.08
		}
	}
	if len(boost) == 0 {
		return nil
	}
	return boost
}

// TechnologyDetection returns technologies, CMS, and confidence for a target.
func (s *Service) TechnologyDetection(ctx context.Context, target string) map[string]any {
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	cms, _ := analysis.Metadata["cms"].(string)
	stackRaw, _ := analysis.Metadata["technology_stack"].([]string)
	if stackRaw == nil {
		if ts, ok := analysis.Metadata["technology_stack"].([]any); ok {
			for _, v := range ts {
				if str, ok := v.(string); ok {
					stackRaw = append(stackRaw, str)
				}
			}
		}
	}
	return map[string]any{
		"target":            analysis.Target,
		"target_type":       analysis.TargetType,
		"technologies":      analysis.Technologies,
		"technology_stack":  stackRaw,
		"cms":               cms,
		"confidence":        analysis.Confidence,
		"risk_level":        analysis.RiskLevel,
	}
}

func (s *Service) candidateIDs(targetType string) []string {
	return s.engine().CandidateTools(targetType)
}

func filterEnabled(names []string, reg *tools.Registry) []string {
	if reg == nil {
		return names
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		spec, ok := reg.Get(name)
		if ok && spec.Enabled {
			out = append(out, name)
		}
	}
	return out
}

func capTools(names []string, objective string) []string {
	switch strings.ToLower(strings.TrimSpace(objective)) {
	case "quick", "fast":
		if len(names) > 3 {
			return names[:3]
		}
	case "focused":
		if len(names) > 5 {
			return names[:5]
		}
	case "stealth":
		if len(names) > 4 {
			return names[:4]
		}
	}
	return names
}

func capToolsWithEngine(names []string, targetType, objective string, eng *DecisionEngine) []string {
	obj := strings.ToLower(strings.TrimSpace(objective))
	if obj == "stealth" {
		ids := make([]string, 0, len(names))
		for _, n := range names {
			ids = append(ids, catalogToShortID(n))
		}
		filtered := filterStealthTools(ids)
		return resolveNames(filtered, names)
	}
	if obj == "comprehensive" && eng != nil {
		ids := make([]string, 0, len(names))
		for _, n := range names {
			ids = append(ids, catalogToShortID(n))
		}
		filtered := filterComprehensiveTools(eng, targetType, ids)
		if len(filtered) > 0 {
			return resolveNames(filtered, names)
		}
	}
	return capTools(names, objective)
}

func catalogToShortID(catalogName string) string {
	for short, full := range tools.BinaryToCatalog {
		if full == catalogName {
			return short
		}
	}
	return catalogName
}

func resolveNames(shortIDs []string, original []string) []string {
	byShort := map[string]string{}
	for _, n := range original {
		byShort[catalogToShortID(n)] = n
	}
	out := make([]string, 0, len(shortIDs))
	seen := map[string]struct{}{}
	for _, id := range shortIDs {
		name, ok := byShort[id]
		if !ok {
			name = id
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func (s *Service) SelectTools(ctx context.Context, targetType, objective string) []string {
	return s.SelectToolsForTarget(ctx, targetType, objective, "")
}

func (s *Service) SelectToolsForTarget(ctx context.Context, targetType, objective, target string) []string {
	cands := s.candidateIDs(targetType)
	_, techLabels, cms, _, _, _ := probeTarget(ctx, target)
	stack := labelsToTechnologies(techLabels, cms)
	boost := mergeBoost(s.graphBoost(ctx, target), techStackBoost(stack), cmsToolBoost(cms, s.Registry))
	ranked := s.engine().RankToolsWithBoost(targetType, cands, boost)
	ranked = appendTechSpecificTools(ranked, stack, cms)
	ranked = s.engine().RankToolsWithBoost(targetType, ranked, boost)
	names := tools.ResolveCatalogNames(ranked, s.Registry)
	names = filterEnabled(names, s.Registry)
	obj := strings.ToLower(strings.TrimSpace(objective))
	if obj == "stealth" {
		ids := ranked
		names = tools.ResolveCatalogNames(filterStealthTools(ids), s.Registry)
		names = filterEnabled(names, s.Registry)
		return capTools(names, objective)
	}
	if obj == "comprehensive" {
		filtered := filterComprehensiveTools(s.engine(), targetType, ranked)
		if len(filtered) > 0 {
			names = tools.ResolveCatalogNames(filtered, s.Registry)
			names = filterEnabled(names, s.Registry)
		}
	}
	if obj == "quick" || obj == "fast" {
		if len(names) > 3 {
			names = names[:3]
		}
		return names
	}
	return capToolsWithEngine(names, targetType, objective, s.engine())
}

func appendTechSpecificTools(ranked []string, stack []Technology, cms string) []string {
	seen := map[string]struct{}{}
	for _, id := range ranked {
		seen[id] = struct{}{}
	}
	add := func(id string) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ranked = append(ranked, id)
	}
	for _, t := range stack {
		switch t {
		case TechWordPress:
			add("wpscan")
		case TechPHP:
			add("nikto")
		}
	}
	if strings.EqualFold(cms, "wordpress") {
		add("wpscan")
	}
	return ranked
}

func labelsToTechnologies(labels []string, cms string) []Technology {
	if cms != "" {
		labels = append(labels, cms)
	}
	seen := map[Technology]struct{}{}
	for _, l := range labels {
		switch strings.ToLower(l) {
		case "wordpress":
			seen[TechWordPress] = struct{}{}
		case "drupal":
			seen[TechDrupal] = struct{}{}
		case "joomla":
			seen[TechJoomla] = struct{}{}
		case "php":
			seen[TechPHP] = struct{}{}
		case "nginx":
			seen[TechNginx] = struct{}{}
		case "apache":
			seen[TechApache] = struct{}{}
		case "nodejs":
			seen[TechNodeJS] = struct{}{}
		case "java":
			seen[TechJava] = struct{}{}
		case "dotnet":
			seen[TechDotNet] = struct{}{}
		}
	}
	out := make([]Technology, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	return out
}

func cmsToolBoost(cms string, reg *tools.Registry) map[string]float64 {
	if cms == "" || reg == nil {
		return nil
	}
	boost := map[string]float64{}
	switch cms {
	case "wordpress":
		boost["wpscan"] = 0.25
		boost["nuclei"] = 0.05
	case "php":
		boost["nikto"] = 0.15
		boost["sqlmap"] = 0.12
	case "drupal", "joomla":
		boost["nuclei"] = 0.1
		boost["nikto"] = 0.1
	}
	for id := range boost {
		name := tools.ResolveCatalogName(id, reg)
		spec, ok := reg.Get(name)
		if !ok || !spec.Enabled {
			delete(boost, id)
		}
	}
	return boost
}

func mergeBoost(parts ...map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for _, p := range parts {
		for k, v := range p {
			out[k] += v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// OptimizeParameters suggests CLI flags for a tool against a target profile.
func (s *Service) OptimizeParameters(targetType, toolName string, params map[string]string) map[string]string {
	return s.OptimizeParametersWithContext(context.Background(), targetType, toolName, params, OptimizeContext{})
}

// OptimizeParametersWithContext applies profile-aware tuning for a target.
func (s *Service) OptimizeParametersWithContext(ctx context.Context, targetType, toolName string, params map[string]string, octx OptimizeContext) map[string]string {
	out := make(map[string]string)
	for k, v := range params {
		out[k] = v
	}
	toolID := toolName
	if s.Registry != nil {
		toolID = tools.ResolveCatalogName(toolName, s.Registry)
	}
	if spec, ok := s.Registry.Get(toolID); ok && spec.Binary != "" {
		toolID = spec.Binary
	}
	target := out["target"]
	if target == "" {
		target = out["url"]
	}
	profile := BuildTargetProfile(target, targetType, nil, "", nil, 0)
	if target != "" {
		profile = s.BuildProfileFromTarget(ctx, target)
		profile.TargetType = targetType
	}
	optimized := s.engine().OptimizeParametersWithProfile(profile, toolID, out, octx)
	for k, v := range optimized {
		out[k] = v
	}
	return out
}

// CreateAttackChain builds an ordered list of catalog tool names from attack patterns.
func (s *Service) CreateAttackChain(ctx context.Context, target string, objective string) map[string]any {
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	profile := s.BuildProfileFromTarget(ctx, target)
	confidence := profile.ConfidenceScore
	if analysis.Confidence > confidence {
		confidence = analysis.Confidence
	}
	octx := OptimizeContext{Objective: objective}
	if strings.EqualFold(objective, "stealth") {
		octx.Stealth = true
	}
	patternKey := SelectPatternKey(analysis.TargetType, objective)
	pattern := AttackPatterns()[patternKey]
	steps := make([]map[string]any, 0, len(pattern))
	var probSum float64
	var timeSum int
	eng := s.engine()
	stepNum := 0
	for _, ps := range pattern {
		catalogName := tools.ResolveCatalogName(ps.Tool, s.Registry)
		spec, ok := s.Registry.Get(catalogName)
		if !ok || !spec.Enabled {
			continue
		}
		score := eng.Score(analysis.TargetType, ps.Tool)
		stepProb := stepSuccessProbability(score, confidence)
		probSum += stepProb
		timeSum += executionTimeEstimate(ps.Tool)
		stepNum++
		params := map[string]string{"target": target}
		for k, v := range ps.Params {
			params[k] = v
		}
		params = s.OptimizeParametersWithContext(ctx, analysis.TargetType, catalogName, params, octx)
		step := map[string]any{
			"step":                     stepNum,
			"tool":                     catalogName,
			"priority":                 ps.Priority,
			"effectiveness_score":      score,
			"success_probability":      stepProb,
			"execution_time_estimate":  executionTimeEstimate(ps.Tool),
			"expected_outcome":         expectedOutcome(ps.Tool),
			"parameters":               params,
		}
		steps = append(steps, step)
	}
	if len(steps) == 0 {
		selected := s.SelectToolsForTarget(ctx, analysis.TargetType, objective, target)
		for i, name := range selected {
			toolID := name
			if spec, ok := s.Registry.Get(name); ok && spec.Binary != "" {
				toolID = spec.Binary
			}
			score := eng.Score(analysis.TargetType, toolID)
			stepProb := stepSuccessProbability(score, confidence)
			probSum += stepProb
			timeSum += executionTimeEstimate(toolID)
			params := s.OptimizeParametersWithContext(ctx, analysis.TargetType, name, map[string]string{"target": target}, octx)
			steps = append(steps, map[string]any{
				"step":                    i + 1,
				"tool":                    name,
				"effectiveness_score":     score,
				"success_probability":     stepProb,
				"execution_time_estimate": executionTimeEstimate(toolID),
				"expected_outcome":        expectedOutcome(toolID),
				"parameters":              params,
			})
		}
		patternKey = "ranked_fallback"
	}
	successProb := 0.0
	if len(steps) > 0 {
		successProb = probSum / float64(len(steps))
	}
	estMinutes := timeSum / 60
	if estMinutes < 1 && len(steps) > 0 {
		estMinutes = 1
	}
	return map[string]any{
		"target":              target,
		"objective":           objective,
		"pattern":             patternKey,
		"analysis":            analysis,
		"steps":               steps,
		"status":              "planned",
		"success_probability": successProb,
		"estimated_minutes":   estMinutes,
		"confidence_score":    confidence,
		"attack_surface_score": profile.AttackSurfaceScore,
	}
}
