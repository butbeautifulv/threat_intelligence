package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/telemetry"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/payloads"
)

type bridgeWorkflowHandler func(ctx context.Context, s *Server, name string, args map[string]any) (any, error)

var bridgeWorkflowHandlers = map[string]bridgeWorkflowHandler{
	"advanced_payload_generation": bridgePayloadStub("advanced_payload_generation"),
	"api_fuzzer": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		target := argTarget(args)
		if s.intel == nil {
			return bridgeOK(name, target, map[string]any{"note": "intelligence not configured; stub fuzz plan only"})
		}
		out := s.intel.ComprehensiveAPIAudit(ctx, "", intelligence.ComprehensiveAPIAuditRequest{
			BaseURL:         firstNonEmpty(argString(args, "base_url", ""), target),
			SchemaURL:       argString(args, "schema_url", ""),
			JWTToken:        argString(args, "jwt_token", ""),
			GraphQLEndpoint: argString(args, "graphql_endpoint", ""),
		})
		return bridgeOK(name, target, map[string]any{"audit": out, "mode": "api_fuzzer"})
	},
	"autorecon_comprehensive": bridgeWorkflowNote("autorecon_comprehensive", "use intelligent_smart_scan or nuclei/ffuf via catalog tools"),
	"autorecon_scan":          bridgeWorkflowNote("autorecon_scan", "use intelligent_smart_scan or subprocess recon tools"),
	"checkov_iac_scan":        bridgeWorkflowNote("checkov_iac_scan", "use checkov CLI when enabled in runner profile"),
	"clair_vulnerability_scan": bridgeWorkflowNote("clair_vulnerability_scan", "use trivy/grype for container scanning"),
	"clear_cache": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		cleared := 0
		if s.runner != nil && s.runner.Cache != nil {
			cleared = s.runner.Cache.Clear()
		}
		return bridgeOK(name, argTarget(args), map[string]any{"cleared": cleared})
	},
	"cloudmapper_analysis": bridgeWorkflowNote("cloudmapper_analysis", "use cloud security catalog tools and graph intel"),
	"create_file": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		if s.files == nil {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": "files manager not configured"})
		}
		fn := argString(args, "filename", argString(args, "file", ""))
		content := argString(args, "content", "")
		out, err := s.files.Create(fn, content, argBool(args, "binary"))
		if err != nil {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": err.Error()})
		}
		out["success"] = true
		return bridgeOK(name, argTarget(args), out)
	},
	"create_scan_summary": bridgeWorkflowNote("create_scan_summary", "use create_vulnerability_report after findings aggregation"),
	"create_vulnerability_report": bridgeWorkflowNote("create_vulnerability_report", "use format_tool_output_visual or HTTP /api/reports"),
	"execute_command": bridgeWorkflowNote("execute_command", "use POST /api/command or catalog subprocess tools"),
	"execute_python_script": bridgeWorkflowNote("execute_python_script", "use execute_command with python3 or dedicated runner tool"),
	"generate_payload": bridgePayloadGenerate,
	"get_cache_stats": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		stats := map[string]any{"entries": 0, "ttl_sec": 0}
		if s.runner != nil && s.runner.Cache != nil {
			stats = s.runner.Cache.Stats()
		}
		stats["success"] = true
		return bridgeOK(name, argTarget(args), stats)
	},
	"get_process_status": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		pid := argInt(args, "pid", 0)
		if s.processes == nil {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": "process manager not configured"})
		}
		if pid == 0 {
			return bridgeOK(name, argTarget(args), map[string]any{"success": true, "processes": s.processes.List()})
		}
		rec, ok := s.processes.Get(pid)
		if !ok {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": fmt.Sprintf("pid %d not found", pid)})
		}
		return bridgeOK(name, argTarget(args), map[string]any{"success": true, "process": rec})
	},
	"get_telemetry": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		_ = args
		out := map[string]any{"success": true, "prometheus": "/metrics", "note": "use GET /api/telemetry for expanded stats"}
		if s.processes != nil {
			out["processes_total"] = len(s.processes.List())
		}
		if s.runner != nil && s.runner.Cache != nil {
			stats := s.runner.Cache.Stats()
			out["cache"] = stats
			if entries, ok := stats["entries"].(int); ok {
				telemetry.SetCacheEntries(entries)
			}
		}
		return bridgeOK(name, argTarget(args), out)
	},
	"http_framework_test":  bridgeHTTPStub("http_framework_test"),
	"http_intruder":        bridgeHTTPStub("http_intruder"),
	"http_repeater":        bridgeHTTPStub("http_repeater"),
	"http_set_rules":       bridgeHTTPStub("http_set_rules"),
	"http_set_scope":       bridgeHTTPStub("http_set_scope"),
	"kube_bench_cis":       bridgeWorkflowNote("kube_bench_cis", "use kube-bench in runner when enabled"),
	"kube_hunter_scan":     bridgeWorkflowNote("kube_hunter_scan", "use kube-hunter in runner when enabled"),
	"list_active_processes": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		if s.processes == nil {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": "process manager not configured"})
		}
		return bridgeOK(name, argTarget(args), map[string]any{"success": true, "processes": s.processes.List()})
	},
	"list_files": func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		if s.files == nil {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": "files manager not configured"})
		}
		out, err := s.files.List(argString(args, "directory", "."))
		if err != nil {
			return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": err.Error()})
		}
		out["success"] = true
		return bridgeOK(name, argTarget(args), out)
	},
}

func (s *Server) tryBridgeWorkflowTool(ctx context.Context, name string, args map[string]any) (any, bool, error) {
	h, ok := bridgeWorkflowHandlers[name]
	if !ok {
		return nil, false, nil
	}
	out, err := h(ctx, s, name, args)
	return out, true, err
}

func bridgeOK(tool, target string, fields map[string]any) (any, error) {
	if fields == nil {
		fields = map[string]any{}
	}
	if _, ok := fields["success"]; !ok {
		fields["success"] = true
	}
	fields["tool"] = tool
	if target != "" {
		fields["target"] = target
	}
	return toolJSONResult(fields)
}

func bridgeWorkflowNote(tool, note string) bridgeWorkflowHandler {
	return func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		_ = s
		return bridgeOK(tool, argTarget(args), map[string]any{"note": note})
	}
}

func bridgePayloadStub(tool string) bridgeWorkflowHandler {
	return func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		if s.files == nil {
			return bridgeOK(tool, argTarget(args), map[string]any{
				"success": false,
				"error":   "files manager not configured",
				"note":    "use ai_generate_payload when files manager is wired",
			})
		}
		return bridgePayloadGenerate(ctx, s, name, args)
	}
}

func bridgePayloadGenerate(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
	_ = ctx
	if s.files == nil {
		return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": "files manager not configured"})
	}
	size := argInt(args, "size", 256)
	if size <= 0 {
		size = 256
	}
	out, err := payloads.Generate(s.files, payloads.Request{
		Type:     argString(args, "type", "buffer"),
		Size:     size,
		Pattern:  argString(args, "pattern", "A"),
		Filename: argString(args, "filename", ""),
	})
	if err != nil {
		return bridgeOK(name, argTarget(args), map[string]any{"success": false, "error": err.Error()})
	}
	out["success"] = true
	out["note"] = "deterministic payload generation (not LLM)"
	return bridgeOK(name, argTarget(args), out)
}

func bridgeHTTPStub(tool string) bridgeWorkflowHandler {
	return func(ctx context.Context, s *Server, name string, args map[string]any) (any, error) {
		_ = ctx
		_ = s
		return bridgeOK(tool, argTarget(args), map[string]any{
			"note":    "use httpx/curl catalog tools or discovery browser for live HTTP testing",
			"method":  argString(args, "method", "GET"),
			"url":     firstNonEmpty(argString(args, "url", ""), argTarget(args)),
			"stub":    true,
			"ts":      time.Now().UTC().Format(time.RFC3339),
			"headers": args["headers"],
			"body":    args["body"],
		})
	}
}
