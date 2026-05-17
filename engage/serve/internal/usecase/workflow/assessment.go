package workflow

import (
	"context"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/report"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// AssessmentReport runs smart-scan and returns scan output plus summary report.
func (s *Service) AssessmentReport(ctx context.Context, subject string, req SmartScanRequest) map[string]any {
	scan := s.SmartScan(ctx, subject, req)
	target, _ := scan["target"].(string)
	var all []domainreport.Finding
	if raw, ok := scan["findings"].([]domainreport.Finding); ok {
		all = raw
	} else if executed, ok := scan["tools_executed"].([]map[string]any); ok {
		for _, e := range executed {
			toolName, _ := e["tool"].(string)
			stdout, _ := e["stdout"].(string)
			all = append(all, findings.ParseToolOutput(toolName, target, stdout)...)
		}
		scan["findings"] = all
		scan["total_vulnerabilities"] = len(all)
		if s.Findings != nil {
			for _, f := range all {
				_ = s.Findings.PublishFinding(ctx, f.Tool, f.Target, f.Title, string(f.Severity), f.Description)
			}
		}
	}
	analysis := s.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	summary := report.FromSmartScan(target, scan)
	exec := report.BuildExecutiveSummary(target, scan, all, analysis.RiskLevel, analysis.Technologies)
	return map[string]any{
		"scan":                scan,
		"summary_report":      summary,
		"findings":            all,
		"severity_breakdown":  report.SeverityBreakdown(all),
		"executive_summary":   exec,
	}
}
