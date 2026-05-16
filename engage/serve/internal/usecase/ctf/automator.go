package ctf

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

var flagPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)flag\{[^}]+\}`),
	regexp.MustCompile(`(?i)ctf\{[^}]+\}`),
	regexp.MustCompile(`[0-9a-f]{32}`),
	regexp.MustCompile(`[0-9a-f]{40}`),
	regexp.MustCompile(`[0-9a-f]{64}`),
}

// AutoSolveOptions configures automated solving.
type AutoSolveOptions struct {
	ExecuteTools bool
	MaxSteps     int
}

// StepResult is the outcome of one workflow step execution.
type StepResult struct {
	Step        int      `json:"step"`
	Action      string   `json:"action"`
	Success     bool     `json:"success"`
	Output      string   `json:"output"`
	ToolsUsed   []string `json:"tools_used"`
	ExecutionMS int64    `json:"execution_time_ms"`
}

// SolveResult is the auto-solve response payload.
type SolveResult struct {
	ChallengeID     string       `json:"challenge_id"`
	Status            string       `json:"status"`
	AutomatedSteps    []StepResult `json:"automated_steps"`
	ManualSteps       []map[string]string `json:"manual_steps"`
	Confidence        float64      `json:"confidence"`
	FlagCandidates    []string     `json:"flag_candidates"`
	NextActions       []string     `json:"next_actions,omitempty"`
	Flag              string       `json:"flag,omitempty"`
}

// Automator runs CTF workflow steps with optional tool execution.
type Automator struct {
	Manager *Manager
	Tools   *toolsuc.Runner
	Registry *tools.Registry
}

// AutoSolve attempts to solve a challenge using workflow steps.
func (a *Automator) AutoSolve(ctx context.Context, subject string, ch Challenge, opts AutoSolveOptions) SolveResult {
	result := SolveResult{
		ChallengeID:  ch.Name,
		Status:       "in_progress",
		AutomatedSteps: []StepResult{},
		ManualSteps:  []map[string]string{},
		FlagCandidates: []string{},
	}
	if opts.MaxSteps <= 0 {
		opts.MaxSteps = 8
	}
	suggested := a.Manager.Tools.SuggestTools(ch.Description, ch.Category)
	suggested = MergePatternTools(ch, suggested)
	resolved := a.Manager.Tools.ResolveTools(suggested, a.Registry)
	wf := a.Manager.CreateChallengeWorkflow(ch, resolved)

	target := ch.TargetOrURL()
	confidence := 0.0
	stepsRun := 0
	var allOutput strings.Builder

	for _, step := range wf.WorkflowSteps {
		if stepsRun >= opts.MaxSteps {
			break
		}
		sr := StepResult{Step: step.Step, Action: step.Action}
		if opts.ExecuteTools && a.Tools != nil && target != "" {
			for _, toolID := range step.Tools {
				if toolID == "manual" || toolID == "custom" {
					continue
				}
				name := tools.ResolveCatalogName(toolID, a.Registry)
				if _, err := a.Registry.MustGet(name); err != nil {
					continue
				}
				start := time.Now()
				res := a.Tools.Run(ctx, subject, name, contract.ToolRunRequest{Target: target})
				sr.ToolsUsed = append(sr.ToolsUsed, name)
				sr.ExecutionMS += time.Since(start).Milliseconds()
				if res.Output != "" {
					allOutput.WriteString(res.Output)
					allOutput.WriteByte('\n')
				}
				if res.Success {
					sr.Success = true
				}
			}
		} else {
			for _, toolID := range step.Tools {
				if toolID != "manual" {
					sr.ToolsUsed = append(sr.ToolsUsed, toolID)
				}
			}
			sr.Output = "[planned] " + step.Description
			sr.Success = true
		}
		result.AutomatedSteps = append(result.AutomatedSteps, sr)
		stepsRun++
		if sr.Success {
			confidence += 0.1
		}
	}

	result.FlagCandidates = extractFlagCandidates(allOutput.String())
	if len(result.FlagCandidates) > 0 && validateFlagFormat(result.FlagCandidates[0]) {
		result.Status = "solved"
		result.Flag = result.FlagCandidates[0]
	} else {
		result.Status = "needs_manual_intervention"
		result.ManualSteps = manualGuidance(ch, result)
	}
	if confidence > 1.0 {
		confidence = 1.0
	}
	result.Confidence = confidence
	return result
}

func extractFlagCandidates(output string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, re := range flagPatterns {
		for _, m := range re.FindAllString(output, -1) {
			if _, dup := seen[m]; dup {
				continue
			}
			seen[m] = struct{}{}
			out = append(out, m)
		}
	}
	return out
}

func validateFlagFormat(flag string) bool {
	patterns := []string{`(?i)^flag\{.+}$`, `(?i)^ctf\{.+}$`, `(?i)^[a-z0-9_]+\{.+}$`}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, flag); matched {
			return true
		}
	}
	return false
}

func manualGuidance(ch Challenge, result SolveResult) []map[string]string {
	var guidance []map[string]string
	switch ch.Category {
	case "web":
		guidance = append(guidance,
			map[string]string{"action": "manual_source_review", "description": "Review HTML/JS source for hidden clues"},
			map[string]string{"action": "parameter_fuzzing", "description": "Fuzz parameters with custom payloads"},
		)
	case "crypto":
		guidance = append(guidance,
			map[string]string{"action": "cipher_research", "description": "Research cipher type and known attacks"},
		)
	case "pwn":
		guidance = append(guidance,
			map[string]string{"action": "manual_debugging", "description": "Debug binary to understand control flow"},
		)
	case "forensics":
		guidance = append(guidance,
			map[string]string{"action": "steganography_deep_dive", "description": "Deep steganography analysis"},
		)
	default:
		guidance = append(guidance,
			map[string]string{"action": "retry_with_tools", "description": "Try alternative tools from suggest-tools"},
		)
	}
	return guidance
}
