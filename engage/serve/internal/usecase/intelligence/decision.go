package intelligence

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
	table, ok := d.effectiveness[targetType]
	if !ok {
		table = d.effectiveness["unknown"]
	}
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

// RankToolsWithBoost applies optional score boosts (e.g. from veil graph context).
func (d *DecisionEngine) RankToolsWithBoost(targetType string, candidates []string, boost map[string]float64) []string {
	table, ok := d.effectiveness[targetType]
	if !ok {
		table = d.effectiveness["unknown"]
	}
	type scored struct {
		id    string
		score float64
	}
	var list []scored
	for _, id := range candidates {
		score := table[id]
		if score == 0 {
			score = 0.5
		}
		if boost != nil {
			score += boost[id]
		}
		list = append(list, scored{id, score})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = s.id
	}
	return out
}

// Score returns effectiveness for a tool against a target type.
func (d *DecisionEngine) Score(targetType, toolID string) float64 {
	table, ok := d.effectiveness[targetType]
	if !ok {
		table = d.effectiveness["unknown"]
	}
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
