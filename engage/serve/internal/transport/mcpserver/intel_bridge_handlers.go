package mcpserver

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

type intelBridgeHandler func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error)

var intelBridgeHandlers = map[string]intelBridgeHandler{
	"analyze_target_intelligence": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = args
		out := s.intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
		return toolJSONResult(out)
	},
	"create_attack_chain_ai": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		obj := argString(args, "objective", "comprehensive")
		out := s.intel.CreateAttackChain(ctx, target, obj)
		return toolJSONResult(out)
	},
	"intelligent_smart_scan": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
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
	},
	"comprehensive_api_audit": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		out := s.intel.ComprehensiveAPIAudit(ctx, subject, intelligence.ComprehensiveAPIAuditRequest{
			BaseURL:         firstNonEmpty(argString(args, "base_url", ""), target),
			SchemaURL:       argString(args, "schema_url", ""),
			JWTToken:        argString(args, "jwt_token", ""),
			GraphQLEndpoint: argString(args, "graphql_endpoint", ""),
		})
		return toolJSONResult(out)
	},
	"api_schema_analyzer": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		url := firstNonEmpty(argString(args, "schema_url", ""), target)
		out := map[string]any{"schema_url": url, "note": "use comprehensive_api_audit with schema_url for full audit"}
		if url != "" {
			out["analysis"] = s.intel.ComprehensiveAPIAudit(ctx, subject, intelligence.ComprehensiveAPIAuditRequest{
				BaseURL:   target,
				SchemaURL: url,
			})
		}
		return toolJSONResult(out)
	},
	"jwt_analyzer": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = ctx
		_ = subject
		_ = target
		_ = spec
		tok := argString(args, "jwt_token", "")
		if tok == "" {
			tok = argString(args, "token", "")
		}
		return toolJSONResult(intelligence.JWTAnalysis(tok))
	},
	"correlate_threat_intelligence": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return toolJSONResult(s.intel.CorrelateThreatIntelligence(ctx, target, argString(args, "indicators", "")))
	},
	"target_graph_context": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return toolJSONResult(s.intel.TargetGraph(ctx, target, argString(args, "indicators", "")))
	},
	"target_timeline_intelligence": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return toolJSONResult(s.intel.TargetTimeline(ctx, intelligence.TargetTimelineRequest{
			Target:       target,
			Limit:        argInt(args, "limit", 50),
			IncludeGraph: argString(args, "include_graph", "true") != "false",
		}))
	},
	"discover_attack_chains": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return toolJSONResult(s.intel.DiscoverAttackChains(ctx, target, argString(args, "objective", "comprehensive")))
	},
	"ai_vulnerability_assessment": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		return toolJSONResult(s.intel.AIVulnerabilityAssessment(ctx, subject, target, argInt(args, "max_tools", 6)))
	},
	"vulnerability_intelligence_dashboard": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = args
		_ = spec
		analysis := s.intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
		return toolJSONResult(map[string]any{
			"target":       target,
			"risk_level":   analysis.RiskLevel,
			"technologies": analysis.Technologies,
			"confidence":   analysis.Confidence,
			"metadata":     analysis.Metadata,
			"dashboard":    "summary",
			"success":      true,
		})
	},
	"bugbounty_reconnaissance_workflow": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "reconnaissance", target)
	},
	"bugbounty_vulnerability_hunting": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "vuln-hunt", target)
	},
	"bugbounty_business_logic_testing": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "business-logic", target)
	},
	"bugbounty_osint_gathering": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "osint", target)
	},
	"bugbounty_file_upload_testing": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "file-upload", target)
	},
	"bugbounty_comprehensive_assessment": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "comprehensive", target)
	},
	"bugbounty_authentication_bypass_testing": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return s.callBugbountyWorkflow(ctx, subject, "business-logic", target)
	},
	"run_playbook": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		return s.callPlaybook(ctx, subject, argString(args, "playbook", argString(args, "name", "")), target, argBool(args, "async"))
	},
	"checksec_analyze": func(ctx context.Context, s *Server, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = ctx
		_ = subject
		_ = spec
		return toolJSONResult(map[string]any{
			"success":     true,
			"target":      target,
			"binary_path": argString(args, "binary_path", ""),
			"note":        "use checksec via runner when binary is in engage-runner image",
		})
	},
}
