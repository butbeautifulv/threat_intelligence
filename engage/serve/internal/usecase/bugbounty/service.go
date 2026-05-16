package bugbounty

import (
	"context"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
)

// Service is the bug bounty usecase facade.
type Service struct {
	Registry *tools.Registry
	Intel    *intelligence.Service
	Manager  *Manager
	Runner   *PhaseRunner
	Jobs     *jobuc.Queue
}

// NewService wires bug bounty dependencies.
func NewService(reg *tools.Registry, intel *intelligence.Service, executor PhaseExecutor, findings FindingPublisher, jobs *jobuc.Queue) *Service {
	var runner *PhaseRunner
	if executor != nil {
		runner = &PhaseRunner{
			Registry: reg,
			Intel:    intel,
			WF:       executor,
			Findings: findings,
		}
	}
	return &Service{
		Registry: reg,
		Intel:    intel,
		Manager:  NewManager(),
		Runner:   runner,
		Jobs:     jobs,
	}
}

func (s *Service) now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Run executes a named bug bounty workflow (plan or plan+execute).
func (s *Service) Run(ctx context.Context, subject, workflowName string, t Target, opts RunOptions) map[string]any {
	switch workflowName {
	case "reconnaissance":
		return s.runRecon(ctx, subject, t, opts)
	case "vuln-hunt":
		return s.runVulnHunt(ctx, subject, t, opts)
	case "business-logic":
		return s.runBusinessLogic(ctx, t, opts)
	case "osint":
		return s.runOSINT(ctx, t, opts)
	case "file-upload":
		return s.runFileUpload(ctx, subject, t, opts)
	case "comprehensive":
		return s.runComprehensive(ctx, subject, t, opts)
	default:
		return map[string]any{"success": false, "error": "unknown workflow: " + workflowName}
	}
}

func (s *Service) runRecon(ctx context.Context, subject string, t Target, opts RunOptions) map[string]any {
	wf := s.Manager.CreateReconnaissance(t)
	out := map[string]any{
		"success":   true,
		"workflow":  wf,
		"timestamp": s.now(),
	}
	if opts.Execute && s.Runner != nil {
		phaseResults, findings := s.Runner.ExecuteRecon(ctx, subject, t, wf)
		out["phase_results"] = phaseResults
		out["findings"] = findings
		out["total_vulnerabilities"] = len(findings)
	}
	if opts.Async && s.Jobs != nil {
		out["async_jobs"] = s.enqueueReconJobs(subject, t, wf, opts.MaxTools)
	}
	return out
}

func (s *Service) runVulnHunt(ctx context.Context, subject string, t Target, opts RunOptions) map[string]any {
	wf := s.Manager.CreateVulnHunt(t)
	out := map[string]any{
		"success":   true,
		"workflow":  wf,
		"timestamp": s.now(),
	}
	if opts.Execute && s.Runner != nil {
		phaseResults, findings := s.Runner.ExecuteVulnHunt(ctx, subject, t, wf)
		out["phase_results"] = phaseResults
		out["findings"] = findings
	}
	return out
}

func (s *Service) runBusinessLogic(_ context.Context, t Target, _ RunOptions) map[string]any {
	wf := s.Manager.CreateBusinessLogic(t)
	return map[string]any{
		"success":   true,
		"workflow":  wf,
		"timestamp": s.now(),
	}
}

func (s *Service) runOSINT(_ context.Context, t Target, _ RunOptions) map[string]any {
	wf := s.Manager.CreateOSINT(t)
	return map[string]any{
		"success":   true,
		"workflow":  wf,
		"timestamp": s.now(),
	}
}

func (s *Service) runFileUpload(_ context.Context, _ string, t Target, _ RunOptions) map[string]any {
	url := t.Domain
	wf := s.Manager.CreateFileUpload(url)
	out := map[string]any{
		"success":   true,
		"workflow":  wf,
		"timestamp": s.now(),
	}
	return out
}

func (s *Service) runComprehensive(ctx context.Context, subject string, t Target, opts RunOptions) map[string]any {
	assessment := s.Manager.CreateComprehensive(t)
	out := map[string]any{
		"success":    true,
		"assessment": assessment,
		"timestamp":  s.now(),
	}
	if opts.Execute && s.Runner != nil {
		recon := s.Manager.CreateReconnaissance(t)
		phaseResults, findings := s.Runner.ExecuteRecon(ctx, subject, t, recon)
		out["phase_results"] = phaseResults
		out["findings"] = findings
	}
	return out
}

func (s *Service) enqueueReconJobs(subject string, t Target, wf ReconWorkflow, maxTools int) []map[string]any {
	if maxTools <= 0 {
		maxTools = 8
	}
	var queued []map[string]any
	count := 0
	for _, phase := range wf.Phases {
		for _, tool := range phase.Tools {
			if count >= maxTools {
				return queued
			}
			name := tools.ResolveCatalogName(tool.Tool, s.Registry)
			if _, err := s.Registry.MustGet(name); err != nil {
				continue
			}
			j, err := s.Jobs.Enqueue(name, t.Domain, subject, tool.Params)
			entry := map[string]any{"tool": name, "phase": phase.Name, "status": "queued"}
			if err != nil {
				entry["status"] = "failed"
				entry["error"] = err.Error()
			} else {
				entry["job_id"] = j.ID
			}
			queued = append(queued, entry)
			count++
		}
	}
	return queued
}

// RunPlaybook runs a bug bounty playbook by workflow name.
func (s *Service) RunPlaybook(ctx context.Context, subject, playbookName, workflowName, target string, async bool, maxTools int) map[string]any {
	t := Target{Domain: target, PriorityVulns: []string{"rce", "sqli", "xss", "idor", "ssrf"}}
	wfName := workflowName
	if wfName == "" {
		wfName = playbookName
	}
	opts := RunOptions{Execute: !async, Async: async, MaxTools: maxTools}
	out := s.Run(ctx, subject, wfName, t, opts)
	out["playbook"] = playbookName
	return out
}

// RunFromBody is the HTTP entrypoint with full JSON body.
func (s *Service) RunFromBody(ctx context.Context, subject, workflowName string, body map[string]any) map[string]any {
	t := TargetFromBody(body)
	opts := RunOptions{
		Execute:  body["execute"] == true,
		Async:    body["async"] == true,
		MaxTools: intVal(body, "max_tools", 8),
	}
	return s.Run(ctx, subject, workflowName, t, opts)
}

func intVal(m map[string]any, k string, def int) int {
	switch v := m[k].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return def
	}
}
