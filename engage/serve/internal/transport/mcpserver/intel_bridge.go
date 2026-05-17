package mcpserver

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/pkg/auth"
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
