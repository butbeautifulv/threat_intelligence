package ctf

import (
	"context"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
)

// Service is the CTF usecase facade for HTTP/MCP handlers.
type Service struct {
	Registry  *tools.Registry
	Tools     *toolsuc.Runner
	Intel     *intelligence.Service
	FilesDir  string
	Manager   *Manager
	Automator *Automator
	Coordinator *Coordinator
	ToolMgr   *ToolManager
}

// NewService wires CTF dependencies.
func NewService(reg *tools.Registry, runner *toolsuc.Runner, intel *intelligence.Service, filesDir string) *Service {
	mgr := NewManager()
	return &Service{
		Registry:    reg,
		Tools:       runner,
		Intel:       intel,
		FilesDir:    filesDir,
		Manager:     mgr,
		ToolMgr:     mgr.Tools,
		Automator:   &Automator{Manager: mgr, Tools: runner, Registry: reg},
		Coordinator: NewCoordinator(),
	}
}

func (s *Service) now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// CreateChallengeWorkflow builds workflow JSON for a challenge.
func (s *Service) CreateChallengeWorkflow(ch Challenge) (map[string]any, error) {
	if err := ch.Validate(true); err != nil {
		return nil, err
	}
	suggested := s.ToolMgr.SuggestTools(ch.Description, ch.Category)
	suggested = MergePatternTools(ch, suggested)
	resolved := s.ToolMgr.ResolveTools(suggested, s.Registry)
	wf := s.Manager.CreateChallengeWorkflow(ch, resolved)
	return map[string]any{
		"success":   true,
		"workflow":  wf,
		"challenge": ch,
		"timestamp": s.now(),
	}, nil
}

// AutoSolve runs the automator.
func (s *Service) AutoSolve(ctx context.Context, subject string, ch Challenge, executeTools bool, maxSteps int) (map[string]any, error) {
	if err := ch.Validate(true); err != nil {
		return nil, err
	}
	result := s.Automator.AutoSolve(ctx, subject, ch, AutoSolveOptions{
		ExecuteTools: executeTools,
		MaxSteps:     maxSteps,
	})
	return map[string]any{
		"success":      true,
		"solve_result": result,
		"challenge":    ch,
		"timestamp":    s.now(),
	}, nil
}

// SuggestTools returns tool suggestions and command templates.
func (s *Service) SuggestTools(description, category, target string) map[string]any {
	suggested := s.ToolMgr.SuggestTools(description, category)
	resolved := s.ToolMgr.ResolveTools(suggested, s.Registry)
	cmds := map[string]string{}
	for _, id := range suggested {
		cmds[id] = s.ToolMgr.ToolCommand(id, target)
	}
	return map[string]any{
		"success":         true,
		"suggested_tools": resolved,
		"raw_tools":       suggested,
		"category_tools":  s.ToolMgr.CategoryToolsFlat(category),
		"tool_commands":   cmds,
		"category":        NormalizeCategory(category),
		"timestamp":       s.now(),
	}
}

// TeamStrategy builds team assignments.
func (s *Service) TeamStrategy(challenges []Challenge, teamSkills map[string][]string) map[string]any {
	strategy := s.Coordinator.TeamStrategy(challenges, teamSkills)
	return map[string]any{
		"success":          true,
		"strategy":         strategy,
		"challenges_count": len(challenges),
		"team_size":        len(teamSkills),
		"timestamp":        s.now(),
	}
}

// AnalyzeCrypto wraps crypto heuristics.
func (s *Service) AnalyzeCrypto(cipherText, cipherType, keyHint, knownPlaintext, additionalInfo string) map[string]any {
	analysis := AnalyzeCrypto(cipherText, cipherType, keyHint, knownPlaintext, additionalInfo)
	return map[string]any{
		"success":   true,
		"analysis":  analysis,
		"timestamp": s.now(),
	}
}

// AnalyzeForensics runs forensics analyzer.
func (s *Service) AnalyzeForensics(ctx context.Context, subject, filePath string, opts ForensicsOptions) map[string]any {
	opts.FilesDir = s.FilesDir
	analysis := AnalyzeForensics(ctx, subject, filePath, opts, s.Tools, s.Registry)
	return map[string]any{
		"success":   true,
		"analysis":  analysis,
		"timestamp": s.now(),
	}
}

// AnalyzeBinary runs binary analyzer.
func (s *Service) AnalyzeBinary(ctx context.Context, subject, binaryPath string, opts BinaryOptions) map[string]any {
	opts.FilesDir = s.FilesDir
	analysis := AnalyzeBinary(ctx, subject, binaryPath, opts, s.Tools, s.Registry)
	return map[string]any{
		"success":   true,
		"analysis":  analysis,
		"timestamp": s.now(),
	}
}

// RunPlaybook runs a CTF playbook (create workflow + optional tool execution).
func (s *Service) RunPlaybook(ctx context.Context, subject string, pb workflow.Playbook, target string, execute bool) map[string]any {
	ch := Challenge{
		Name:        pb.Name,
		Category:    pb.Objective,
		Description: pb.Description,
		Target:      target,
		Difficulty:  "unknown",
	}
	if ch.Category == "ctf" {
		ch.Category = "web"
	}
	if pb.Workflow == "ctf-pwn" {
		ch.Category = "pwn"
	}
	out, err := s.CreateChallengeWorkflow(ch)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	if execute && target != "" {
		solve, _ := s.AutoSolve(ctx, subject, ch, true, pb.MaxTools)
		out["solve_result"] = solve["solve_result"]
	}
	out["playbook"] = pb.Name
	return out
}

// ChallengeFromBody parses challenge fields from HTTP JSON body.
func ChallengeFromBody(body map[string]any) Challenge {
	ch := Challenge{
		Name:        str(body, "name"),
		Category:    str(body, "category"),
		Description: str(body, "description"),
		Difficulty:  str(body, "difficulty"),
		Points:      intVal(body, "points", 100),
		URL:         str(body, "url"),
		Target:      str(body, "target"),
	}
	if ch.Target == "" {
		ch.Target = ch.URL
	}
	if files, ok := body["files"].([]any); ok {
		for _, f := range files {
			if s, ok := f.(string); ok {
				ch.Files = append(ch.Files, s)
			}
		}
	}
	return ch
}

func str(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
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
