package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

// IsIntelBridgeTool returns true when tools/call should use in-process intelligence handlers.
func IsIntelBridgeTool(name string, spec tool.Spec) bool {
	if name == "comprehensive_api_audit" || name == "target_timeline_intelligence" || name == "target_graph_context" {
		return true
	}
	if name == "monitor_cve_feeds" || name == "generate_exploit_from_cve" {
		return true
	}
	if spec.Category == toolid.CategoryCTF {
		return true
	}
	if strings.HasPrefix(name, "ctf_") {
		return true
	}
	return spec.Category == toolid.CategoryIntel
}

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
}

type cveBridgeHandler func(ctx context.Context, s *Server, args map[string]any) (any, error)

var cveBridgeHandlers = map[string]cveBridgeHandler{
	"monitor_cve_feeds": func(ctx context.Context, s *Server, args map[string]any) (any, error) {
		return toolJSONResult(s.cve.MonitorFromBody(ctx, args))
	},
	"generate_exploit_from_cve": func(ctx context.Context, s *Server, args map[string]any) (any, error) {
		return toolJSONResult(s.cve.GenerateExploitFromCVE(ctx, args))
	},
}

type ctfBridgeHandler func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error)

var ctfBridgeHandlers = map[string]ctfBridgeHandler{
	"ctf_create_challenge_workflow": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		ch := ctf.ChallengeFromBody(args)
		ch.Name = firstNonEmpty(ch.Name, target, "challenge")
		ch.Target = firstNonEmpty(ch.Target, target)
		out, err := s.ctf.CreateChallengeWorkflow(ch)
		if err != nil {
			return nil, rpcErrf(codeToolError, "%v", err)
		}
		return toolJSONResult(out)
	},
	"ctf_auto_solve_challenge": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		ch := ctf.ChallengeFromBody(args)
		ch.Name = firstNonEmpty(ch.Name, target, "challenge")
		ch.Target = firstNonEmpty(ch.Target, target)
		exec := argString(args, "execute_tools", "true") != "false"
		out, err := s.ctf.AutoSolve(ctx, subject, ch, exec, argInt(args, "max_steps", 8))
		if err != nil {
			return nil, rpcErrf(codeToolError, "%v", err)
		}
		return toolJSONResult(out)
	},
	"ctf_suggest_tools": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = ctx
		_ = subject
		desc := argString(args, "description", "")
		if desc == "" {
			return nil, rpcErrf(codeToolError, "description required")
		}
		return toolJSONResult(s.ctf.SuggestTools(desc, argString(args, "category", "misc"), target))
	},
	"ctf_team_strategy": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = ctx
		_ = subject
		_ = target
		_ = args
		return toolJSONResult(map[string]any{
			"success": true,
			"note":    "use HTTP POST /api/ctf/team-strategy with challenges[] and team_skills",
		})
	},
	"ctf_cryptography_solver": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = target
		text := argString(args, "cipher_text", "")
		if text == "" {
			return nil, rpcErrf(codeToolError, "cipher_text required")
		}
		return toolJSONResult(s.ctf.AnalyzeCrypto(text, argString(args, "cipher_type", "unknown"),
			argString(args, "key_hint", ""), argString(args, "known_plaintext", ""),
			argString(args, "additional_info", "")))
	},
	"ctf_forensics_analyzer": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = target
		path := argString(args, "file_path", "")
		if path == "" {
			return nil, rpcErrf(codeToolError, "file_path required")
		}
		return toolJSONResult(s.ctf.AnalyzeForensics(ctx, subject, path, ctf.ForensicsOptions{}))
	},
	"ctf_binary_analyzer": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = target
		path := argString(args, "binary_path", "")
		if path == "" {
			return nil, rpcErrf(codeToolError, "binary_path required")
		}
		return toolJSONResult(s.ctf.AnalyzeBinary(ctx, subject, path, ctf.BinaryOptions{}))
	},
}

func (s *Server) callIntelBridge(ctx context.Context, name string, spec tool.Spec, args map[string]any) (any, error) {
	subject := ""
	if sub, ok := auth.SubjectFromContext(ctx); ok {
		subject = sub.Sub
	}
	target := argTarget(args)

	if strings.HasPrefix(name, "ctf_") {
		return s.callCTFBridge(ctx, name, subject, target, args)
	}
	if name == "monitor_cve_feeds" || name == "generate_exploit_from_cve" {
		return s.callCVEBridge(ctx, name, args)
	}
	if s.intel == nil {
		return nil, rpcErrf(codeToolError, "intelligence service not configured")
	}

	if h, ok := intelBridgeHandlers[name]; ok {
		return h(ctx, s, subject, target, args, spec)
	}

	if out, ok, err := s.tryPlaybookByName(ctx, subject, name, target, argBool(args, "async")); ok {
		return out, err
	}

	return toolJSONResult(map[string]any{
		"tool":     name,
		"target":   target,
		"success":  false,
		"error":    "intelligence tool not mapped; use HTTP /api/intelligence/* or enable subprocess binary",
		"category": string(spec.Category),
	})
}

func (s *Server) callCTFBridge(ctx context.Context, name, subject, target string, args map[string]any) (map[string]any, error) {
	if s.ctf == nil {
		return nil, rpcErrf(codeToolError, "ctf service not configured")
	}
	if h, ok := ctfBridgeHandlers[name]; ok {
		return h(ctx, s, subject, target, args)
	}
	return nil, rpcErrf(codeMethodNotFound, "unknown ctf tool: %s", name)
}

func (s *Server) callBugbountyWorkflow(ctx context.Context, subject, wf, target string) (any, error) {
	body := map[string]any{"domain": target, "target": target}
	if s.bugbounty != nil {
		return toolJSONResult(s.bugbounty.RunFromBody(ctx, subject, wf, body))
	}
	if s.workflows == nil {
		return nil, rpcErrf(codeToolError, "workflow service not configured")
	}
	return toolJSONResult(s.workflows.RunWorkflowWithBody(ctx, subject, wf, body))
}

func (s *Server) callPlaybook(ctx context.Context, subject, name, target string, async bool) (any, error) {
	if s.workflows == nil {
		return nil, rpcErrf(codeToolError, "workflow service not configured")
	}
	list, err := workflow.LoadAllPlaybooks(s.catalogPath)
	if err != nil {
		return nil, rpcErrf(codeToolError, "playbooks: %v", err)
	}
	pb, ok := workflow.FindPlaybook(list, name)
	if !ok {
		return nil, rpcErrf(codeToolError, "playbook not found: %s", name)
	}
	if strings.HasPrefix(pb.Workflow, "ctf-") && s.ctf != nil {
		return toolJSONResult(s.ctf.RunPlaybook(ctx, subject, pb, target, !async))
	}
	if isBugBountyPlaybookName(pb.Workflow, pb.Name) && s.bugbounty != nil {
		return toolJSONResult(s.bugbounty.RunPlaybook(ctx, subject, pb.Name, pb.Workflow, target, async, pb.MaxTools))
	}
	return toolJSONResult(s.workflows.RunPlaybook(ctx, subject, pb, target, async))
}

func (s *Server) tryPlaybookByName(ctx context.Context, subject, name, target string, async bool) (any, bool, error) {
	if s.workflows == nil || s.catalogPath == "" {
		return nil, false, nil
	}
	list, err := workflow.LoadAllPlaybooks(s.catalogPath)
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

func (s *Server) callCVEBridge(ctx context.Context, name string, args map[string]any) (any, error) {
	if s.cve == nil {
		return nil, rpcErrf(codeToolError, "CVE service not configured")
	}
	if h, ok := cveBridgeHandlers[name]; ok {
		return h(ctx, s, args)
	}
	return nil, rpcErrf(codeToolError, "unknown CVE tool %q", name)
}

func isBugBountyPlaybookName(workflow, name string) bool {
	switch workflow {
	case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
		return true
	}
	switch name {
	case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
		return true
	}
	return false
}
