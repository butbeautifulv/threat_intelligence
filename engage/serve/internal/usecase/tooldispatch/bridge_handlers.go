package tooldispatch

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

type intelBridgeHandler func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error)

func intelBridgeStub(toolName, note string) intelBridgeHandler {
	return func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = ctx
		_ = d
		_ = subject
		_ = spec
		return map[string]any{
			"tool":    toolName,
			"target":  target,
			"success": true,
			"note":    note,
		}, nil
	}
}

var intelBridgeHandlers = map[string]intelBridgeHandler{
	"analyze_target_intelligence": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = args
		_ = spec
		return d.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target}), nil
	},
	"create_attack_chain_ai": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		obj := argString(args, "objective", "comprehensive")
		out := d.Intel.CreateAttackChain(ctx, target, obj)
		if _, ok := out["success"]; !ok {
			out["success"] = true
		}
		return out, nil
	},
	"intelligent_smart_scan": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		if d.Workflows == nil {
			return nil, dispatchToolError("workflow service not configured")
		}
		return d.Workflows.SmartScan(ctx, subject, workflow.SmartScanRequest{
			Target:    target,
			Objective: argString(args, "objective", "comprehensive"),
			MaxTools:  argInt(args, "max_tools", 5),
			Async:     argBool(args, "async"),
		}), nil
	},
	"comprehensive_api_audit": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		out := d.Intel.ComprehensiveAPIAudit(ctx, subject, intelligence.ComprehensiveAPIAuditRequest{
			BaseURL:         firstNonEmpty(argString(args, "base_url", ""), "http://"+target),
			SchemaURL:       argString(args, "schema_url", ""),
			JWTToken:        argString(args, "jwt_token", ""),
			GraphQLEndpoint: argString(args, "graphql_endpoint", ""),
		})
		if s, ok := out["success"].(bool); !ok || !s {
			out["success"] = true
			out["note"] = "audit dispatched; some phases may be empty without runner tools"
		}
		return out, nil
	},
	"objdump_analyze": intelBridgeStub("objdump_analyze", "use gdb_analyze or radare2_analyze in runner-full"),
	"volatility_analyze": intelBridgeStub("volatility_analyze", "use volatility3_analyze in runner-full"),
	"volatility3_analyze": intelBridgeStub("volatility3_analyze", "use volatility3 wrapper in runner-full"),
	"api_schema_analyzer": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		url := firstNonEmpty(argString(args, "schema_url", ""), target)
		out := map[string]any{
			"schema_url": url,
			"success":    true,
			"note":       "use comprehensive_api_audit with schema_url for full audit",
		}
		if url != "" {
			out["analysis"] = d.Intel.ComprehensiveAPIAudit(ctx, subject, intelligence.ComprehensiveAPIAuditRequest{
				BaseURL:   firstNonEmpty(argString(args, "base_url", ""), "http://"+target),
				SchemaURL: url,
			})
		}
		return out, nil
	},
	"jwt_analyzer": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = ctx
		_ = subject
		_ = target
		_ = spec
		tok := argString(args, "jwt_token", "")
		if tok == "" {
			tok = argString(args, "token", "")
		}
		return intelligence.JWTAnalysis(tok), nil
	},
	"correlate_threat_intelligence": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return d.Intel.CorrelateThreatIntelligence(ctx, target, argString(args, "indicators", "")), nil
	},
	"target_graph_context": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return d.Intel.TargetGraph(ctx, target, argString(args, "indicators", "")), nil
	},
	"target_timeline_intelligence": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return d.Intel.TargetTimeline(ctx, intelligence.TargetTimelineRequest{
			Target:       target,
			Limit:        argInt(args, "limit", 50),
			IncludeGraph: argString(args, "include_graph", "true") != "false",
		}), nil
	},
	"discover_attack_chains": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = spec
		return d.Intel.DiscoverAttackChains(ctx, target, argString(args, "objective", "comprehensive")), nil
	},
	"ai_vulnerability_assessment": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		return d.Intel.AIVulnerabilityAssessment(ctx, subject, target, argInt(args, "max_tools", 6)), nil
	},
	"vulnerability_intelligence_dashboard": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = subject
		_ = args
		_ = spec
		analysis := d.Intel.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
		return map[string]any{
			"target":       target,
			"risk_level":   analysis.RiskLevel,
			"technologies": analysis.Technologies,
			"confidence":   analysis.Confidence,
			"metadata":     analysis.Metadata,
			"dashboard":    "summary",
			"success":      true,
		}, nil
	},
	"bugbounty_reconnaissance_workflow": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "reconnaissance", target)
	},
	"bugbounty_vulnerability_hunting": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "vuln-hunt", target)
	},
	"bugbounty_business_logic_testing": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "business-logic", target)
	},
	"bugbounty_osint_gathering": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "osint", target)
	},
	"bugbounty_file_upload_testing": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "file-upload", target)
	},
	"bugbounty_comprehensive_assessment": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "comprehensive", target)
	},
	"bugbounty_authentication_bypass_testing": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = args
		_ = spec
		return d.callBugbountyWorkflow(ctx, subject, "business-logic", target)
	},
	"run_playbook": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any, spec tool.Spec) (any, error) {
		_ = spec
		return d.callPlaybook(ctx, subject, argString(args, "playbook", argString(args, "name", "")), target, argBool(args, "async"))
	},
}
