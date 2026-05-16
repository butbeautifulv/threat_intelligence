package mcpserver

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/payloads"
)

func (s *Server) tryAgentTool(ctx context.Context, name string, args map[string]any) (any, bool, error) {
	switch name {
	case "ai_generate_payload":
		if s.files == nil {
			out, err := toolJSONResult(map[string]any{"success": false, "error": "files manager not configured"})
			return out, true, err
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
			return nil, true, rpcErrf(codeToolError, "%v", err)
		}
		out["note"] = "deterministic payload generation (not LLM)"
		res, err := toolJSONResult(out)
		return res, true, err

	case "ai_generate_attack_suite":
		if s.intel == nil {
			res, err := toolJSONResult(map[string]any{"success": false, "error": "intelligence not configured"})
			return res, true, err
		}
		target := argTarget(args)
		objective := argString(args, "objective", "comprehensive")
		chain := s.intel.CreateAttackChain(ctx, target, objective)
		out := map[string]any{
			"target":       target,
			"objective":    objective,
			"attack_chain": chain,
			"success":      true,
			"note":         "deterministic attack chain from patterns + ranked tools (not LLM)",
		}
		if s.files != nil {
			if p, err := payloads.Generate(s.files, payloads.Request{Type: "buffer", Size: 64, Pattern: "A"}); err == nil {
				out["sample_payload"] = p
			}
		}
		res, err := toolJSONResult(out)
		return res, true, err

	case "browser_agent_inspect":
		if s.browser == nil || !s.browser.Enabled() {
			res, err := toolJSONResult(map[string]any{
				"success": false,
				"error":   "browser sidecar not configured (ENGAGE_BROWSER_URL)",
			})
			return res, true, err
		}
		target := argTarget(args)
		params := map[string]string{}
		for k, v := range args {
			if s, ok := v.(string); ok {
				params[k] = s
			}
		}
		out := s.browser.Inspect(ctx, browser.InspectFromParams(target, params))
		res, err := toolJSONResult(out)
		return res, true, err

	case "get_process_dashboard", "get_live_dashboard":
		if s.processes == nil {
			res, err := toolJSONResult(map[string]any{"success": false, "error": "process manager not configured"})
			return res, true, err
		}
		res, err := toolJSONResult(s.processes.Dashboard())
		return res, true, err

	case "format_tool_output_visual":
		toolName := argString(args, "tool_name", argString(args, "tool", ""))
		output := argString(args, "output", "")
		target := argTarget(args)
		parsed := findings.ParseToolOutput(toolName, target, output)
		severity := map[string]int{}
		for _, f := range parsed {
			severity[string(f.Severity)]++
		}
		res, err := toolJSONResult(map[string]any{
			"tool":               toolName,
			"target":             target,
			"findings_count":     len(parsed),
			"severity_breakdown": severity,
			"findings":           parsed,
			"visual":             "structured_json",
			"success":            true,
		})
		return res, true, err
	}
	if strings.HasPrefix(name, "ai_generate_") && name != "ai_generate_payload" && name != "ai_generate_attack_suite" {
		res, err := toolJSONResult(map[string]any{
			"tool":    name,
			"success": false,
			"note":    "use ai_generate_payload or HTTP /api/payloads/generate; not an LLM",
		})
		return res, true, err
	}
	return nil, false, nil
}
