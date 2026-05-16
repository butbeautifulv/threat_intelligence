package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

// IsIntelBridgeTool returns true when tools/call should use in-process intelligence handlers.
func IsIntelBridgeTool(name string, spec tool.Spec) bool {
	if name == "comprehensive_api_audit" {
		return true
	}
	return spec.Category == toolid.CategoryIntel
}

func (s *Server) callIntelBridge(ctx context.Context, name string, spec tool.Spec, args map[string]any) (any, error) {
	if s.intel == nil {
		return nil, rpcErrf(codeToolError, "intelligence service not configured")
	}
	subject := ""
	if sub, ok := auth.SubjectFromContext(ctx); ok {
		subject = sub.Sub
	}
	target := argTarget(args)

	switch name {
	case "analyze_target_intelligence":
		out := s.intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
		return toolJSONResult(out)

	case "create_attack_chain_ai":
		obj := argString(args, "objective", "comprehensive")
		out := s.intel.CreateAttackChain(ctx, target, obj)
		return toolJSONResult(out)

	case "intelligent_smart_scan":
		if s.workflows == nil {
			return nil, rpcErrf(codeToolError, "workflow service not configured")
		}
		out := s.workflows.SmartScan(ctx, subject, workflow.SmartScanRequest{
			Target:    target,
			Objective: argString(args, "objective", "comprehensive"),
			MaxTools:  argInt(args, "max_tools", 5),
			Async:     argBool(args, "async"),
		})
		return toolJSONResult(out)

	case "comprehensive_api_audit":
		out := s.intel.ComprehensiveAPIAudit(ctx, subject, intelligence.ComprehensiveAPIAuditRequest{
			BaseURL:         firstNonEmpty(argString(args, "base_url", ""), target),
			SchemaURL:       argString(args, "schema_url", ""),
			JWTToken:        argString(args, "jwt_token", ""),
			GraphQLEndpoint: argString(args, "graphql_endpoint", ""),
		})
		return toolJSONResult(out)

	case "api_schema_analyzer":
		url := firstNonEmpty(argString(args, "schema_url", ""), target)
		out := map[string]any{"schema_url": url, "note": "use comprehensive_api_audit with schema_url for full audit"}
		if url != "" {
			out["analysis"] = s.intel.ComprehensiveAPIAudit(ctx, subject, intelligence.ComprehensiveAPIAuditRequest{
				BaseURL:   target,
				SchemaURL: url,
			})
		}
		return toolJSONResult(out)

	case "jwt_analyzer":
		tok := argString(args, "jwt_token", "")
		if tok == "" {
			tok = argString(args, "token", "")
		}
		return toolJSONResult(intelligence.JWTAnalysis(tok))

	case "correlate_threat_intelligence":
		return toolJSONResult(s.intel.CorrelateThreatIntelligence(ctx, target, argString(args, "indicators", "")))

	case "discover_attack_chains":
		return toolJSONResult(s.intel.DiscoverAttackChains(ctx, target, argString(args, "objective", "comprehensive")))

	case "ai_vulnerability_assessment":
		return toolJSONResult(s.intel.AIVulnerabilityAssessment(ctx, subject, target, argInt(args, "max_tools", 6)))

	case "vulnerability_intelligence_dashboard":
		analysis := s.intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
		return toolJSONResult(map[string]any{
			"target":        target,
			"risk_level":    analysis.RiskLevel,
			"technologies":  analysis.Technologies,
			"confidence":    analysis.Confidence,
			"metadata":      analysis.Metadata,
			"dashboard":     "summary",
			"success":       true,
		})

	case "bugbounty_reconnaissance_workflow":
		return s.callBugbountyWorkflow(ctx, subject, "reconnaissance", target)
	case "bugbounty_vulnerability_hunting":
		return s.callBugbountyWorkflow(ctx, subject, "vuln-hunt", target)
	case "bugbounty_business_logic_testing":
		return s.callBugbountyWorkflow(ctx, subject, "business-logic", target)
	case "bugbounty_osint_gathering":
		return s.callBugbountyWorkflow(ctx, subject, "osint", target)
	case "bugbounty_file_upload_testing":
		return s.callBugbountyWorkflow(ctx, subject, "file-upload", target)
	case "bugbounty_comprehensive_assessment":
		return s.callBugbountyWorkflow(ctx, subject, "comprehensive", target)
	case "bugbounty_authentication_bypass_testing":
		return s.callBugbountyWorkflow(ctx, subject, "business-logic", target)

	case "run_playbook":
		return s.callPlaybook(ctx, subject, argString(args, "playbook", argString(args, "name", "")), target, argBool(args, "async"))
	}

	if out, ok, err := s.tryPlaybookByName(ctx, subject, name, target, argBool(args, "async")); ok {
		return out, err
	}

	// Unknown intelligence tool: return guidance instead of bogus subprocess.
	return toolJSONResult(map[string]any{
		"tool":    name,
		"target":  target,
		"success": false,
		"error":   "intelligence tool not mapped; use HTTP /api/intelligence/* or enable subprocess binary",
		"category": string(spec.Category),
	})
}

func (s *Server) callBugbountyWorkflow(ctx context.Context, subject, wf, target string) (any, error) {
	if s.workflows == nil {
		return nil, rpcErrf(codeToolError, "workflow service not configured")
	}
	return toolJSONResult(s.workflows.RunWorkflow(ctx, subject, wf, target))
}

func (s *Server) callPlaybook(ctx context.Context, subject, name, target string, async bool) (any, error) {
	if s.workflows == nil {
		return nil, rpcErrf(codeToolError, "workflow service not configured")
	}
	list, err := workflow.LoadPlaybooks(workflow.DefaultPlaybooksPath(s.catalogPath))
	if err != nil {
		return nil, rpcErrf(codeToolError, "playbooks: %v", err)
	}
	pb, ok := workflow.FindPlaybook(list, name)
	if !ok {
		return nil, rpcErrf(codeToolError, "playbook not found: %s", name)
	}
	return toolJSONResult(s.workflows.RunPlaybook(ctx, subject, pb, target, async))
}

func (s *Server) tryPlaybookByName(ctx context.Context, subject, name, target string, async bool) (any, bool, error) {
	if s.workflows == nil || s.catalogPath == "" {
		return nil, false, nil
	}
	list, err := workflow.LoadPlaybooks(workflow.DefaultPlaybooksPath(s.catalogPath))
	if err != nil || len(list) == 0 {
		return nil, false, nil
	}
	if _, ok := workflow.FindPlaybook(list, name); !ok {
		return nil, false, nil
	}
	out, err := s.callPlaybook(ctx, subject, name, target, async)
	return out, true, err
}

func toolJSONResult(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(b)},
		},
	}, nil
}

func argTarget(args map[string]any) string {
	if args == nil {
		return ""
	}
	for _, k := range []string{"target", "base_url", "url", "domain", "host"} {
		if v, ok := args[k]; ok {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return s
			}
		}
	}
	return ""
}

func argString(args map[string]any, key, def string) string {
	if args == nil {
		return def
	}
	if v, ok := args[key]; ok {
		if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
			return s
		}
	}
	return def
}

func argInt(args map[string]any, key string, def int) int {
	if args == nil {
		return def
	}
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return def
	}
}

func argBool(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	switch v := args[key].(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	}
	return false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
