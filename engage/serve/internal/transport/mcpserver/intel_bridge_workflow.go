package mcpserver

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
)

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
