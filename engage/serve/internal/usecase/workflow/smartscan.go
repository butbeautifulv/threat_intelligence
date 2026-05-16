package workflow

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	domainreport "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// SmartScanRequest configures intelligent multi-tool execution.
type SmartScanRequest struct {
	Target    string
	Objective string
	MaxTools  int
	Async     bool
}

// SmartScan runs ranked tools against a target (parallel sync or async jobs).
func (s *Service) SmartScan(ctx context.Context, subject string, req SmartScanRequest) map[string]any {
	target := req.Target
	maxTools := req.MaxTools
	if maxTools <= 0 {
		maxTools = 5
	}
	analysis := s.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	selected := s.Intel.SelectToolsForTarget(ctx, analysis.TargetType, req.Objective, target)
	if len(selected) > maxTools {
		selected = selected[:maxTools]
	}

	out := map[string]any{
		"target":         target,
		"objective":      req.Objective,
		"target_profile": analysis,
		"tools_selected": selected,
		"async":          req.Async,
	}

	if len(selected) == 0 {
		out["tools_executed"] = []any{}
		out["findings"] = []domainreport.Finding{}
		out["total_vulnerabilities"] = 0
		out["status"] = "no_tools"
		return out
	}

	if req.Async && s.Jobs != nil {
		executed := make([]map[string]any, 0, len(selected))
		for _, toolName := range selected {
			params := s.Intel.OptimizeParameters(analysis.TargetType, toolName, map[string]string{"target": target})
			j, err := s.Jobs.Enqueue(toolName, target, subject, params)
			entry := map[string]any{
				"tool":       toolName,
				"parameters": params,
				"status":     "queued",
			}
			if err != nil {
				entry["status"] = "failed"
				entry["error"] = err.Error()
			} else {
				entry["job_id"] = j.ID
			}
			executed = append(executed, entry)
		}
		out["tools_executed"] = executed
		out["status"] = "queued"
		return out
	}

	executed := s.runToolsParallel(ctx, subject, target, analysis.TargetType, selected)
	out["tools_executed"] = executed
	allFindings := aggregateFindings(executed, target)
	out["findings"] = allFindings
	out["total_vulnerabilities"] = len(allFindings)
	out["status"] = "completed"
	if s.Findings != nil {
		for _, f := range allFindings {
			_ = s.Findings.PublishFinding(ctx, f.Tool, f.Target, f.Title, string(f.Severity), f.Description)
		}
	}
	return out
}

func aggregateFindings(executed []map[string]any, target string) []domainreport.Finding {
	var all []domainreport.Finding
	for _, e := range executed {
		toolName, _ := e["tool"].(string)
		stdout, _ := e["stdout"].(string)
		all = append(all, findings.ParseToolOutput(toolName, target, stdout)...)
	}
	return all
}

func (s *Service) runToolsParallel(ctx context.Context, subject, target, targetType string, toolNames []string) []map[string]any {
	if s.Tools == nil {
		return nil
	}
	const workers = 5
	sem := make(chan struct{}, workers)
	var mu sync.Mutex
	results := make([]map[string]any, 0, len(toolNames))
	var wg sync.WaitGroup

	for _, name := range toolNames {
		toolName := tools.ResolveCatalogName(name, s.Tools.Registry)
		wg.Add(1)
		go func(catalogName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			params := s.Intel.OptimizeParameters(targetType, catalogName, map[string]string{"target": target})
			start := time.Now()
			res := s.Tools.Run(ctx, subject, catalogName, contract.ToolRunRequest{
				Target:     target,
				Parameters: params,
			})
			toolFindings := findings.ParseToolOutput(catalogName, target, res.Output)
			entry := map[string]any{
				"tool":                  catalogName,
				"parameters":            params,
				"status":                "success",
				"success":               res.Success,
				"execution_time":        time.Since(start).Seconds(),
				"stdout":                res.Output,
				"error":                 res.Error,
				"findings":              toolFindings,
				"vulnerabilities_found": len(toolFindings),
			}
			if !res.Success {
				entry["status"] = "failed"
			}
			if strings.Contains(res.Output, "[recovery:") {
				entry["recovered"] = true
			}
			mu.Lock()
			results = append(results, entry)
			mu.Unlock()
		}(toolName)
	}
	wg.Wait()
	return results
}
