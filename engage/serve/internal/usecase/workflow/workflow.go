package workflow

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// Service runs multi-step workflows (bug bounty, assessment).
type Service struct {
	Intel *intelligence.Service
	Tools *toolsuc.Runner
}

func (s *Service) RunWorkflow(ctx context.Context, subject, name string, target string) map[string]any {
	analysis := s.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	selected := s.Intel.SelectTools(ctx, analysis.TargetType, "comprehensive")
	results := make([]contract.ToolRunResponse, 0, len(selected))
	for _, toolName := range selected {
		if s.Tools == nil {
			break
		}
		resolved := tools.ResolveCatalogName(toolName, s.Tools.Registry)
		if _, err := s.Tools.Registry.MustGet(resolved); err != nil {
			continue
		}
		results = append(results, s.Tools.Run(ctx, subject, resolved, contract.ToolRunRequest{Target: target}))
	}
	return map[string]any{
		"workflow": name,
		"target":   target,
		"analysis": analysis,
		"tools":    selected,
		"results":  results,
	}
}

func (s *Service) Reconnaissance(ctx context.Context, subject, target string) map[string]any {
	return s.RunWorkflow(ctx, subject, "bugbounty-reconnaissance", target)
}

func (s *Service) Comprehensive(ctx context.Context, subject, target string) map[string]any {
	return s.RunWorkflow(ctx, subject, "comprehensive-assessment", target)
}

// SummaryReport builds a minimal JSON report.
func SummaryReport(target string, data map[string]any) json.RawMessage {
	b, _ := json.Marshal(map[string]any{
		"report_type": "summary",
		"target":      target,
		"sections":    data,
	})
	return b
}
