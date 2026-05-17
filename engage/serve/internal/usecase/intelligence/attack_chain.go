package intelligence

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

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
			"step":                    stepNum,
			"tool":                    catalogName,
			"priority":                ps.Priority,
			"effectiveness_score":     score,
			"success_probability":     stepProb,
			"execution_time_estimate": executionTimeEstimate(ps.Tool),
			"expected_outcome":        expectedOutcome(ps.Tool),
			"parameters":              params,
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
				"step":                      i + 1,
				"tool":                      name,
				"effectiveness_score":       score,
				"success_probability":       stepProb,
				"execution_time_estimate": executionTimeEstimate(toolID),
				"expected_outcome":          expectedOutcome(toolID),
				"parameters":                params,
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
		"target":               target,
		"objective":            objective,
		"pattern":              patternKey,
		"analysis":             analysis,
		"steps":                steps,
		"status":               "planned",
		"success_probability":  successProb,
		"estimated_minutes":    estMinutes,
		"confidence_score":     confidence,
		"attack_surface_score": profile.AttackSurfaceScore,
	}
}
