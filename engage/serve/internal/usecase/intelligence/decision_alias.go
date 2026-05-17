package intelligence

import "github.com/butbeautifulv/veil/pkg/decision"

type (
	DecisionEngine  = decision.DecisionEngine
	TargetProfile   = decision.TargetProfile
	OptimizeContext = decision.OptimizeContext
	AttackStep      = decision.AttackStep
)

var (
	DefaultDecisionEngine = decision.DefaultDecisionEngine
	AttackPatterns        = decision.AttackPatterns
	SelectPatternKey      = decision.SelectPatternKey
	BuildTargetProfile    = decision.BuildTargetProfile
)

func stepSuccessProbability(effectiveness, confidence float64) float64 {
	return decision.StepSuccessProbability(effectiveness, confidence)
}

func executionTimeEstimate(toolID string) int {
	return decision.ExecutionTimeEstimate(toolID)
}

func expectedOutcome(toolID string) string {
	return decision.ExpectedOutcome(toolID)
}

func filterStealthTools(toolIDs []string) []string {
	return decision.FilterStealthTools(toolIDs)
}

func filterComprehensiveTools(eng *DecisionEngine, targetType string, toolIDs []string) []string {
	return decision.FilterComprehensiveTools(eng, targetType, toolIDs)
}

func capTools(names []string, objective string) []string {
	return decision.CapTools(names, objective)
}

func capToolsWithEngine(names []string, targetType, objective string, eng *DecisionEngine) []string {
	return decision.CapToolsWithEngine(names, targetType, objective, eng, catalogToShortID)
}
