package workflow

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/bugbounty"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/visual"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
)

// Service runs multi-step workflows (bug bounty, assessment).
type Service struct {
	Intel      *intelligence.Service
	Tools      *toolsuc.Runner
	Jobs       *jobuc.Queue
	Findings   FindingBus
	BugBounty   *bugbounty.Service
	Progress    *visual.Store
	MaxParallel int
}

func (s *Service) maxParallel() int {
	if s.MaxParallel > 0 {
		return s.MaxParallel
	}
	return 5
}

func (s *Service) RunWorkflow(ctx context.Context, subject, name string, target string) map[string]any {
	return s.RunWorkflowWithBody(ctx, subject, name, map[string]any{"target": target, "domain": target})
}

// RunWorkflowWithBody runs a workflow using a full JSON body (domain, execute, etc.).
func (s *Service) RunWorkflowWithBody(ctx context.Context, subject, name string, body map[string]any) map[string]any {
	if body == nil {
		body = map[string]any{}
	}
	if s.BugBounty != nil {
		switch name {
		case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
			return s.BugBounty.RunFromBody(ctx, subject, name, body)
		}
	}
	if name == "comprehensive" {
		target, _ := body["target"].(string)
		if target == "" {
			target, _ = body["domain"].(string)
		}
		return s.Comprehensive(ctx, subject, target)
	}
	target, _ := body["target"].(string)
	if target == "" {
		target, _ = body["domain"].(string)
	}
	if s.BugBounty != nil {
		return s.BugBounty.RunFromBody(ctx, subject, name, body)
	}
	return map[string]any{"success": false, "workflow": name, "target": target}
}

func (s *Service) Reconnaissance(ctx context.Context, subject, target string) map[string]any {
	return s.RunWorkflow(ctx, subject, "bugbounty-reconnaissance", target)
}

func (s *Service) Comprehensive(ctx context.Context, subject, target string) map[string]any {
	return s.AssessmentReport(ctx, subject, SmartScanRequest{
		Target:    target,
		Objective: "comprehensive",
		MaxTools:  8,
		Async:     false,
	})
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
