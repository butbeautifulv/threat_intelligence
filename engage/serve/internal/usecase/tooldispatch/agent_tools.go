package tooldispatch

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/findings"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/payloads"
	pkgreport "github.com/butbeautifulv/veil/pkg/report"
)

func (d *Dispatcher) tryAgentTool(ctx context.Context, name string, args map[string]any) (any, bool, error) {
	switch name {
	case "ai_generate_payload":
		if d.Files == nil {
			return map[string]any{"success": false, "error": "files manager not configured"}, true, nil
		}
		size := argInt(args, "size", 256)
		if size <= 0 {
			size = 256
		}
		out, err := payloads.Generate(d.Files, payloads.Request{
			Type:     argString(args, "type", "buffer"),
			Size:     size,
			Pattern:  argString(args, "pattern", "A"),
			Filename: argString(args, "filename", ""),
		})
		if err != nil {
			return nil, true, dispatchToolError("%v", err)
		}
		out["note"] = "deterministic payload generation (not LLM)"
		return out, true, nil

	case "ai_generate_attack_suite":
		if d.Intel == nil {
			return map[string]any{"success": false, "error": "intelligence not configured"}, true, nil
		}
		target := argTarget(args)
		objective := argString(args, "objective", "comprehensive")
		chain := d.Intel.CreateAttackChain(ctx, target, objective)
		out := map[string]any{
			"target":       target,
			"objective":    objective,
			"attack_chain": chain,
			"success":      true,
			"note":         "deterministic attack chain from patterns + ranked tools (not LLM)",
		}
		if d.Files != nil {
			if p, err := payloads.Generate(d.Files, payloads.Request{Type: "buffer", Size: 64, Pattern: "A"}); err == nil {
				out["sample_payload"] = p
			}
		}
		return out, true, nil

	case "browser_agent_inspect":
		if d.Browser == nil || !d.Browser.Enabled() {
			return map[string]any{
				"success": false,
				"error":   "discovery browser not configured (DISCOVERY_BROWSER_URL)",
			}, true, nil
		}
		target := argTarget(args)
		params := map[string]string{}
		for k, v := range args {
			if s, ok := v.(string); ok {
				params[k] = s
			}
		}
		return d.Browser.Inspect(ctx, browser.InspectFromParams(target, params)), true, nil

	case "get_process_dashboard", "get_live_dashboard":
		if d.Processes == nil {
			return map[string]any{"success": false, "error": "process manager not configured"}, true, nil
		}
		return d.Processes.Dashboard(), true, nil

	case "format_tool_output_visual":
		toolName := argString(args, "tool_name", argString(args, "tool", ""))
		output := argString(args, "output", "")
		target := argTarget(args)
		parsed := findings.ParseToolOutput(toolName, target, output)
		severity := pkgreport.SeverityBreakdown(parsed)
		return map[string]any{
			"tool":               toolName,
			"target":             target,
			"findings_count":     len(parsed),
			"severity_breakdown": severity,
			"findings":           parsed,
			"visual":             "structured_json",
			"success":            true,
		}, true, nil

	case "ai_reconnaissance_workflow", "ai_test_payload":
		return map[string]any{
			"tool":    name,
			"target":  argTarget(args),
			"success": true,
			"note":    "deterministic workflow stub (not LLM); use intelligent_smart_scan or ai_generate_payload",
		}, true, nil
	}
	if strings.HasPrefix(name, "ai_generate_") && name != "ai_generate_payload" && name != "ai_generate_attack_suite" {
		return map[string]any{
			"tool":    name,
			"success": true,
			"note":    "deterministic workflow stub; use ai_generate_payload or HTTP /api/payloads/generate for payloads",
		}, true, nil
	}
	return nil, false, nil
}
