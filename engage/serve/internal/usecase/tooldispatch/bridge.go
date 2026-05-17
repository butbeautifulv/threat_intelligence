package tooldispatch

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
)

func (d *Dispatcher) callIntelBridge(ctx context.Context, name string, spec tool.Spec, args map[string]any) (any, error) {
	subject := ""
	if sub, ok := auth.SubjectFromContext(ctx); ok {
		subject = sub.Sub
	}
	target := argTarget(args)

	if strings.HasPrefix(name, "ctf_") {
		return d.callCTFBridge(ctx, name, subject, target, args)
	}
	if name == "monitor_cve_feeds" || name == "generate_exploit_from_cve" {
		return d.callCVEBridge(ctx, name, args)
	}
	if d.Intel == nil {
		return nil, dispatchToolError("intelligence service not configured")
	}

	if h, ok := intelBridgeHandlers[name]; ok {
		return h(ctx, d, subject, target, args, spec)
	}

	if out, ok, err := d.tryPlaybookByName(ctx, subject, name, target, argBool(args, "async")); ok {
		return out, err
	}

	return map[string]any{
		"tool":     name,
		"target":   target,
		"success":  false,
		"error":    "intelligence tool not mapped; use HTTP /api/intelligence/* or enable subprocess binary",
		"category": string(spec.Category),
	}, nil
}
