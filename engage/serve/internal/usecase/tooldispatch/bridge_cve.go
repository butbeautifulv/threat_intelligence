package tooldispatch

import "context"

func (d *Dispatcher) callCVEBridge(ctx context.Context, name string, args map[string]any) (any, error) {
	if d.CVE == nil {
		return nil, dispatchToolError("CVE service not configured")
	}
	switch name {
	case "monitor_cve_feeds":
		return d.CVE.MonitorFromBody(ctx, args), nil
	case "generate_exploit_from_cve":
		return d.CVE.GenerateExploitFromCVE(ctx, args), nil
	default:
		return nil, dispatchToolError("unknown CVE tool %q", name)
	}
}
