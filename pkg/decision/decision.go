package decision

import (
	"cmp"
	"slices"
)

// DecisionEngine scores tools per target type (port of HexStrike IntelligentDecisionEngine tables).
type DecisionEngine struct {
	effectiveness map[string]map[string]float64
}

func DefaultDecisionEngine() *DecisionEngine {
	tables := defaultEffectivenessTables()
	return &DecisionEngine{effectiveness: tables}
}

// CandidateTools returns all tool ids with effectiveness scores for a target type.
func (d *DecisionEngine) CandidateTools(targetType string) []string {
	table := d.effectivenessTable(targetType)
	out := make([]string, 0, len(table))
	for id := range table {
		out = append(out, id)
	}
	return out
}

// RankTools returns tool ids sorted by effectiveness for targetType.
func (d *DecisionEngine) RankTools(targetType string, candidates []string) []string {
	return d.RankToolsWithBoost(targetType, candidates, nil)
}

type rankedTool struct {
	id    string
	score float64
}

func compareRankedDesc(a, b rankedTool) int { return cmp.Compare(b.score, a.score) }

func boostValue(boost map[string]float64, id string) float64 {
	if boost == nil {
		return 0
	}
	return boost[id]
}

func (d *DecisionEngine) effectivenessTable(targetType string) map[string]float64 {
	if t, ok := d.effectiveness[targetType]; ok {
		return t
	}
	return d.effectiveness["unknown"]
}

// RankToolsWithBoost applies optional score boosts (e.g. from veil graph context).
func (d *DecisionEngine) RankToolsWithBoost(targetType string, candidates []string, boost map[string]float64) []string {
	table := d.effectivenessTable(targetType)
	var list []rankedTool
	for _, id := range candidates {
		score := table[id]
		if score == 0 {
			score = 0.5
		}
		score += boostValue(boost, id)
		list = append(list, rankedTool{id: id, score: score})
	}
	slices.SortFunc(list, compareRankedDesc)
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = s.id
	}
	return out
}

// Score returns effectiveness for a tool against a target type.
func (d *DecisionEngine) Score(targetType, toolID string) float64 {
	table := d.effectivenessTable(targetType)
	if s, ok := table[toolID]; ok {
		return s
	}
	return 0.5
}

// OptimizeParameters applies tool-specific defaults from the decision engine.
func (d *DecisionEngine) OptimizeParameters(targetType, toolID string, params map[string]string) map[string]string {
	p := TargetProfile{TargetType: targetType}
	if params != nil {
		if t, ok := params["target"]; ok {
			p.Target = t
		}
	}
	return d.OptimizeParametersWithProfile(p, toolID, params, OptimizeContext{})
}
