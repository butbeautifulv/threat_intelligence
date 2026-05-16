package intelligence

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ExecuteAttackChain plans and runs enabled pattern steps, passing pattern parameters to the runner.
func (s *Service) ExecuteAttackChain(ctx context.Context, subject, target, objective string) map[string]any {
	chain := s.CreateAttackChain(ctx, target, objective)
	steps, _ := chain["steps"].([]map[string]any)
	executed := make([]map[string]any, 0, len(steps))
	if s.Tools == nil {
		chain["status"] = "planned"
		return chain
	}
	analysis, _ := chain["analysis"].(contract.AnalyzeTargetResponse)
	for _, step := range steps {
		toolName, _ := step["tool"].(string)
		if toolName == "" {
			continue
		}
		params := map[string]string{"target": target}
		if raw, ok := step["parameters"].(map[string]string); ok {
			for k, v := range raw {
				params[k] = v
			}
		}
		optimized := s.OptimizeParameters(analysis.TargetType, toolName, params)
		res := s.Tools.Run(ctx, subject, toolName, contract.ToolRunRequest{
			Target:     target,
			Parameters: optimized,
		})
		executed = append(executed, map[string]any{
			"step":       step["step"],
			"tool":       toolName,
			"parameters": optimized,
			"success":    res.Success,
			"output":     res.Output,
			"error":      res.Error,
		})
	}
	chain["executed"] = executed
	chain["status"] = "executed"
	return chain
}
