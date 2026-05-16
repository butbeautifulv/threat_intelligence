package workflow

import "context"

// RunPlaybook executes a YAML playbook template against a target.
func (s *Service) RunPlaybook(ctx context.Context, subject string, pb Playbook, target string, async bool) map[string]any {
	if pb.MaxTools <= 0 {
		pb.MaxTools = 5
	}
	objective := pb.Objective
	if objective == "" {
		objective = pb.Workflow
	}
	if objective == "" {
		objective = pb.Name
	}
	if pb.Workflow == "comprehensive" || pb.Name == "comprehensive" {
		out := s.Comprehensive(ctx, subject, target)
		out["playbook"] = pb.Name
		return out
	}
	out := s.SmartScan(ctx, subject, SmartScanRequest{
		Target:    target,
		Objective: objective,
		MaxTools:  pb.MaxTools,
		Async:     async,
	})
	out["playbook"] = pb.Name
	out["workflow"] = pb.Workflow
	return out
}
