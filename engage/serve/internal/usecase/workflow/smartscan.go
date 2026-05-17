package workflow

import (
	"context"
	"strings"
	"sync"
	"time"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// SmartScanRequest configures intelligent multi-tool execution.
type SmartScanRequest struct {
	Target            string
	Objective         string
	MaxTools          int
	Async             bool
	RateLimitCheck    bool
	effectiveParallel int // set after rate-limit probe
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

	var scanID string
	if s.Progress != nil && len(selected) > 0 {
		scanID = s.Progress.Create(target, "smart_scan", selected)
	}

	out := map[string]any{
		"target":         target,
		"objective":      req.Objective,
		"target_profile": analysis,
		"tools_selected": selected,
		"async":          req.Async,
	}
	if scanID != "" {
		out["scan_id"] = scanID
	}

	if len(selected) == 0 {
		out["tools_executed"] = []any{}
		out["findings"] = []domainreport.Finding{}
		out["total_vulnerabilities"] = 0
		out["status"] = "no_tools"
		out["success"] = true
		if scanID != "" && s.Progress != nil {
			s.Progress.Finish(scanID, "completed")
		}
		return out
	}

	maxWorkers := s.maxParallel()
	if req.RateLimitCheck && s.Intel != nil {
		probe := s.Intel.ProbeRateLimit(ctx, subject, target)
		out["rate_limit_probe"] = probe
		if probe.Detected {
			if maxWorkers > 2 {
				maxWorkers = 2
			}
			out["recommendation"] = "reduce parallelism due to rate limiting"
		}
	}
	req.effectiveParallel = maxWorkers
	out["max_parallel"] = maxWorkers

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
		out["success"] = true
		return out
	}

	executed := s.runToolsParallel(ctx, subject, target, analysis.TargetType, selected, scanID, req.effectiveParallel)
	out["tools_executed"] = executed
	allFindings := findings.DedupeFindings(aggregateFindings(executed, target))
	out["findings"] = allFindings
	out["total_vulnerabilities"] = len(allFindings)
	out["status"] = "completed"
	if scanID != "" && s.Progress != nil {
		s.Progress.Finish(scanID, "completed")
	}
	if s.Findings != nil {
		for _, f := range allFindings {
			_ = s.Findings.PublishFinding(ctx, f.Tool, f.Target, f.Title, string(f.Severity), f.Description)
		}
	}
	out["success"] = true
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

// RunToolsParallel executes catalog tools concurrently (exported for bugbounty phased runs).
func (s *Service) RunToolsParallel(ctx context.Context, subject, target, targetType string, toolNames []string) []map[string]any {
	return s.runToolsParallel(ctx, subject, target, targetType, toolNames, "", 0)
}

func (s *Service) runToolsParallel(ctx context.Context, subject, target, targetType string, toolNames []string, scanID string, maxWorkers int) []map[string]any {
	if s.Tools == nil {
		return nil
	}
	if maxWorkers <= 0 {
		maxWorkers = s.maxParallel()
	}
	var mu sync.Mutex
	results := make([]map[string]any, 0, len(toolNames))

	RunBounded(maxWorkers, toolNames, func(name string) {
		catalogName := tools.ResolveCatalogName(name, s.Tools.Registry)
		if scanID != "" && s.Progress != nil {
			s.Progress.StartTool(scanID, catalogName)
		}
		params := s.Intel.OptimizeParameters(targetType, catalogName, map[string]string{"target": target})
		start := time.Now()
		res := s.Tools.Run(ctx, subject, catalogName, contract.ToolRunRequest{
			Target:     target,
			Parameters: params,
		})
		elapsed := time.Since(start).Seconds()
		toolFindings := findings.ParseToolOutput(catalogName, target, res.Output)
		st := "success"
		if !res.Success {
			st = "failed"
		}
		if scanID != "" && s.Progress != nil {
			s.Progress.CompleteTool(scanID, catalogName, st, elapsed)
		}
		entry := map[string]any{
			"tool":                  catalogName,
			"parameters":            params,
			"status":                st,
			"success":               res.Success,
			"execution_time":        elapsed,
			"stdout":                res.Output,
			"error":                 res.Error,
			"findings":              toolFindings,
			"vulnerabilities_found": len(toolFindings),
		}
		if strings.Contains(res.Output, "[recovery:") {
			entry["recovered"] = true
		}
		mu.Lock()
		results = append(results, entry)
		mu.Unlock()
	})
	return results
}
