package bugbounty

import (
	"context"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// PhaseRunner executes phased workflows.
type PhaseRunner struct {
	Registry *tools.Registry
	Intel    *intelligence.Service
	WF       PhaseExecutor
	Findings FindingPublisher
}

// PhaseExecutor runs tools in parallel (implemented by workflow.Service).
type PhaseExecutor interface {
	RunToolsParallel(ctx context.Context, subject, target, targetType string, toolNames []string) []map[string]any
}

// FindingPublisher publishes findings to the event bus.
type FindingPublisher interface {
	PublishFinding(ctx context.Context, tool, target, title, severity, description string) error
}

// PhaseResult is the outcome of executing one phase.
type PhaseResult struct {
	Name            string `json:"name"`
	ToolsExecuted   int    `json:"tools_executed"`
	FindingsCount   int    `json:"findings_count"`
	EstimatedTime   int    `json:"estimated_time"`
	Results         []map[string]any `json:"results,omitempty"`
}

// ExecuteRecon runs all recon phases when execute is enabled.
func (p *PhaseRunner) ExecuteRecon(ctx context.Context, subject string, t Target, wf ReconWorkflow) ([]PhaseResult, []domainreport.Finding) {
	var phaseResults []PhaseResult
	var allFindings []domainreport.Finding
	tt := "web"
	if p.Intel != nil {
		analysis := p.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: t.Domain})
		tt = analysis.TargetType
	}
	for _, phase := range wf.Phases {
		toolIDs := ToolIDsFromPhase(phase)
		resolved := tools.ResolveCatalogNames(toolIDs, p.Registry)
		if len(resolved) == 0 || p.WF == nil {
			phaseResults = append(phaseResults, PhaseResult{
				Name: phase.Name, EstimatedTime: phase.EstimatedTime,
			})
			continue
		}
		executed := p.WF.RunToolsParallel(ctx, subject, t.Domain, tt, resolved)
		phaseFindings := aggregatePhaseFindings(executed, t.Domain)
		p.publishFindings(ctx, phaseFindings)
		phaseResults = append(phaseResults, PhaseResult{
			Name:          phase.Name,
			ToolsExecuted: len(executed),
			FindingsCount: len(phaseFindings),
			EstimatedTime: phase.EstimatedTime,
			Results:       executed,
		})
		allFindings = append(allFindings, phaseFindings...)
	}
	return phaseResults, allFindings
}

func aggregatePhaseFindings(executed []map[string]any, target string) []domainreport.Finding {
	var all []domainreport.Finding
	for _, e := range executed {
		toolName, _ := e["tool"].(string)
		stdout, _ := e["stdout"].(string)
		if stdout == "" {
			if o, ok := e["output"].(string); ok {
				stdout = o
			}
		}
		all = append(all, findings.ParseToolOutput(toolName, target, stdout)...)
	}
	return all
}

func (p *PhaseRunner) publishFindings(ctx context.Context, list []domainreport.Finding) {
	if p.Findings == nil {
		return
	}
	for _, f := range list {
		_ = p.Findings.PublishFinding(ctx, f.Tool, f.Target, f.Title, string(f.Severity), f.Description)
	}
}

// ExecuteVulnHunt runs tools for each vulnerability test entry.
func (p *PhaseRunner) ExecuteVulnHunt(ctx context.Context, subject string, t Target, wf VulnHuntWorkflow) ([]PhaseResult, []domainreport.Finding) {
	var phaseResults []PhaseResult
	var allFindings []domainreport.Finding
	tt := "web"
	if p.Intel != nil {
		analysis := p.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: t.Domain})
		tt = analysis.TargetType
	}
	for _, vt := range wf.VulnerabilityTests {
		resolved := tools.ResolveCatalogNames(vt.Tools, p.Registry)
		if len(resolved) == 0 || p.WF == nil {
			continue
		}
		executed := p.WF.RunToolsParallel(ctx, subject, t.Domain, tt, resolved)
		phaseFindings := aggregatePhaseFindings(executed, t.Domain)
		p.publishFindings(ctx, phaseFindings)
		phaseResults = append(phaseResults, PhaseResult{
			Name:          vt.VulnerabilityType,
			ToolsExecuted: len(executed),
			FindingsCount: len(phaseFindings),
			EstimatedTime: vt.EstimatedTime,
		})
		allFindings = append(allFindings, phaseFindings...)
	}
	return phaseResults, allFindings
}
