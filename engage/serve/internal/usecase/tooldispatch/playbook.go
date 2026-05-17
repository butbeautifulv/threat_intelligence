package tooldispatch

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
)

func (d *Dispatcher) callBugbountyWorkflow(ctx context.Context, subject, wf, target string) (any, error) {
	body := map[string]any{"domain": target, "target": target}
	if d.BugBounty != nil {
		return d.BugBounty.RunFromBody(ctx, subject, wf, body), nil
	}
	if d.Workflows == nil {
		return nil, dispatchToolError("workflow service not configured")
	}
	return d.Workflows.RunWorkflowWithBody(ctx, subject, wf, body), nil
}

func (d *Dispatcher) callPlaybook(ctx context.Context, subject, name, target string, async bool) (any, error) {
	if d.Workflows == nil {
		return nil, dispatchToolError("workflow service not configured")
	}
	list, err := workflow.LoadAllPlaybooks(d.CatalogPath)
	if err != nil {
		return nil, dispatchToolError("playbooks: %v", err)
	}
	pb, ok := workflow.FindPlaybook(list, name)
	if !ok {
		return nil, dispatchToolError("playbook not found: %s", name)
	}
	if strings.HasPrefix(pb.Workflow, "ctf-") && d.CTF != nil {
		return d.CTF.RunPlaybook(ctx, subject, pb, target, !async), nil
	}
	if isBugBountyPlaybookName(pb.Workflow, pb.Name) && d.BugBounty != nil {
		return d.BugBounty.RunPlaybook(ctx, subject, pb.Name, pb.Workflow, target, async, pb.MaxTools), nil
	}
	return d.Workflows.RunPlaybook(ctx, subject, pb, target, async), nil
}

func (d *Dispatcher) tryPlaybookByName(ctx context.Context, subject, name, target string, async bool) (any, bool, error) {
	if d.Workflows == nil || d.CatalogPath == "" {
		return nil, false, nil
	}
	list, err := workflow.LoadAllPlaybooks(d.CatalogPath)
	if err != nil || len(list) == 0 {
		return nil, false, nil
	}
	if _, ok := workflow.FindPlaybook(list, name); !ok {
		return nil, false, nil
	}
	out, err := d.callPlaybook(ctx, subject, name, target, async)
	return out, true, err
}

func isBugBountyPlaybookName(workflowName, name string) bool {
	switch workflowName {
	case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
		return true
	}
	switch name {
	case "reconnaissance", "vuln-hunt", "business-logic", "osint", "file-upload", "comprehensive":
		return true
	}
	return false
}
