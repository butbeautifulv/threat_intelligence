package intelligence

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ExecuteAttackChain plans and runs enabled pattern steps, passing pattern parameters to the runner.
func (s *Service) ExecuteAttackChain(ctx context.Context, subject, target, objective string, parallel bool) map[string]any {
	chain := s.CreateAttackChain(ctx, target, objective)
	steps, _ := chain["steps"].([]map[string]any)
	if s.Tools == nil {
		chain["status"] = "planned"
		return chain
	}
	analysis, _ := chain["analysis"].(contract.AnalyzeTargetResponse)
	if parallel && s.ParallelRunner != nil {
		chain["executed"] = s.executeChainParallel(ctx, subject, target, objective, steps, analysis)
		chain["parallel"] = true
	} else {
		chain["executed"] = s.executeChainSequential(ctx, subject, target, objective, steps, analysis)
		chain["parallel"] = false
	}
	chain["status"] = "executed"
	return chain
}

func (s *Service) executeChainSequential(ctx context.Context, subject, target, objective string, steps []map[string]any, analysis contract.AnalyzeTargetResponse) []map[string]any {
	executed := make([]map[string]any, 0, len(steps))
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
		octx := OptimizeContext{Objective: objective}
		optimized := s.OptimizeParametersWithContext(ctx, analysis.TargetType, toolName, params, octx)
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
	return executed
}

func (s *Service) executeChainParallel(ctx context.Context, subject, target, objective string, steps []map[string]any, analysis contract.AnalyzeTargetResponse) []map[string]any {
	toolNames := make([]string, 0, len(steps))
	stepByTool := make(map[string]map[string]any, len(steps))
	for _, step := range steps {
		toolName, _ := step["tool"].(string)
		if toolName == "" {
			continue
		}
		toolNames = append(toolNames, toolName)
		stepByTool[toolName] = step
	}
	results := s.ParallelRunner.RunToolsParallel(ctx, subject, target, analysis.TargetType, toolNames)
	executed := make([]map[string]any, 0, len(results))
	for _, res := range results {
		toolName, _ := res["tool"].(string)
		step := stepByTool[toolName]
		entry := map[string]any{
			"tool":    toolName,
			"success": res["success"],
			"output":  res["stdout"],
			"error":   res["error"],
		}
		if step != nil {
			entry["step"] = step["step"]
			if params, ok := res["parameters"].(map[string]string); ok {
				entry["parameters"] = params
			}
		}
		executed = append(executed, entry)
	}
	return executed
}
