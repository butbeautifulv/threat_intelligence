package ctf

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func ctfGoldenDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "golden")
}

type challengeWorkflowGoldenSpec struct {
	Challenge           string   `json:"challenge"`
	Category            string   `json:"category"`
	Difficulty          string   `json:"difficulty"`
	Points              int      `json:"points"`
	ToolsMin            int      `json:"tools_min"`
	EstimatedTime       int      `json:"estimated_time"`
	SuccessProbability  float64  `json:"success_probability"`
	AutomationLevel     string   `json:"automation_level"`
	StrategyNamesSorted []string `json:"strategy_names_sorted"`
	WorkflowSteps       []struct {
		Step        int      `json:"step"`
		Action      string   `json:"action"`
		Parallel    bool     `json:"parallel"`
		ToolsSorted []string `json:"tools_sorted"`
	} `json:"workflow_steps"`
	ParallelTasksSorted   []string `json:"parallel_tasks_sorted"`
	ValidationStepsSorted []string `json:"validation_steps_sorted"`
}

func assertChallengeWorkflowGolden(t *testing.T, wf ChallengeWorkflow, goldenFile string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(ctfGoldenDir(t), goldenFile))
	if err != nil {
		t.Fatal(err)
	}
	var spec challengeWorkflowGoldenSpec
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if wf.Challenge != spec.Challenge || wf.Category != spec.Category || wf.Difficulty != spec.Difficulty || wf.Points != spec.Points {
		t.Fatalf("metadata mismatch got %+v want challenge=%q category=%q difficulty=%q points=%d", wf, spec.Challenge, spec.Category, spec.Difficulty, spec.Points)
	}
	if len(wf.Tools) < spec.ToolsMin {
		t.Fatalf("tools len %d < %d", len(wf.Tools), spec.ToolsMin)
	}
	if wf.EstimatedTime != spec.EstimatedTime {
		t.Fatalf("estimated_time %d want %d", wf.EstimatedTime, spec.EstimatedTime)
	}
	if math.Abs(wf.SuccessProbability-spec.SuccessProbability) > 1e-9 {
		t.Fatalf("success_probability %v want %v", wf.SuccessProbability, spec.SuccessProbability)
	}
	if wf.AutomationLevel != spec.AutomationLevel {
		t.Fatalf("automation_level %q want %q", wf.AutomationLevel, spec.AutomationLevel)
	}

	gotStrategies := make([]string, len(wf.Strategies))
	for i, s := range wf.Strategies {
		gotStrategies[i] = s.Strategy
	}
	slices.Sort(gotStrategies)
	wantStrategies := append([]string(nil), spec.StrategyNamesSorted...)
	slices.Sort(wantStrategies)
	if diff := cmpSortedSlicesExact(gotStrategies, wantStrategies); diff != "" {
		t.Fatal(diff)
	}

	if len(wf.WorkflowSteps) != len(spec.WorkflowSteps) {
		t.Fatalf("workflow steps len %d want %d", len(wf.WorkflowSteps), len(spec.WorkflowSteps))
	}
	for i, step := range spec.WorkflowSteps {
		got := wf.WorkflowSteps[i]
		if got.Step != step.Step || got.Action != step.Action || got.Parallel != step.Parallel {
			t.Fatalf("step[%d] meta mismatch got %+v want step=%d action=%q parallel=%v", i, got, step.Step, step.Action, step.Parallel)
		}
		if diff := cmpSortedSlicesExact(sortedCopy(got.Tools), sortedCopy(step.ToolsSorted)); diff != "" {
			t.Fatalf("step[%d] tools: %s", i, diff)
		}
	}

	gotPar := sortedCopy(wf.ParallelTasks)
	wantPar := sortedCopy(spec.ParallelTasksSorted)
	if diff := cmpSortedSlicesExact(gotPar, wantPar); diff != "" {
		t.Fatal(diff)
	}

	gotVal := sortedCopy(wf.ValidationSteps)
	wantVal := sortedCopy(spec.ValidationStepsSorted)
	if diff := cmpSortedSlicesExact(gotVal, wantVal); diff != "" {
		t.Fatal(diff)
	}
}

func challengeWorkflowFromGolden(t *testing.T, ch Challenge) ChallengeWorkflow {
	t.Helper()
	mgr := NewManager()
	if err := ch.Validate(false); err != nil {
		t.Fatal(err)
	}
	suggested := mgr.Tools.SuggestTools(ch.Description, ch.Category)
	return mgr.CreateChallengeWorkflow(ch, suggested)
}

func TestGolden_CreateChallengeWeb(t *testing.T) {
	ch := Challenge{Name: "Login Portal", Category: "web", Description: "A basic login page", Difficulty: "medium", Points: 100}
	assertChallengeWorkflowGolden(t, challengeWorkflowFromGolden(t, ch), "create_challenge_web.golden.json")
}

func TestGolden_CreateChallengePwn(t *testing.T) {
	ch := Challenge{Name: "Stack Smash", Category: "pwn", Description: "buffer overflow binary", Difficulty: "hard", Points: 400}
	assertChallengeWorkflowGolden(t, challengeWorkflowFromGolden(t, ch), "create_challenge_pwn.golden.json")
}

func TestGolden_CreateChallengeCrypto(t *testing.T) {
	ch := Challenge{Name: "RSA Vault", Category: "crypto", Description: "RSA public key ciphertext", Difficulty: "medium", Points: 150}
	assertChallengeWorkflowGolden(t, challengeWorkflowFromGolden(t, ch), "create_challenge_crypto.golden.json")
}

func TestGolden_CreateChallengeForensics(t *testing.T) {
	ch := Challenge{Name: "Hidden Image", Category: "forensics", Description: "stego image with hidden flag", Difficulty: "medium", Points: 200}
	assertChallengeWorkflowGolden(t, challengeWorkflowFromGolden(t, ch), "create_challenge_forensics.golden.json")
}

func TestGolden_TeamStrategy(t *testing.T) {
	c := NewCoordinator()
	ts := c.TeamStrategy([]Challenge{
		{Name: "web1", Category: "web", Points: 200, Difficulty: "medium"},
		{Name: "pwn1", Category: "pwn", Points: 300, Difficulty: "medium"},
	}, map[string][]string{"alice": {"web"}, "bob": {"pwn"}})

	b, err := os.ReadFile(filepath.Join(ctfGoldenDir(t), "team_strategy.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		AssignmentsMin     int               `json:"assignments_min"`
		EstimatedTotal     int               `json:"estimated_total"`
		RecommendedOrder   []string          `json:"recommended_order"`
		ChallengeAssignees map[string]string `json:"challenge_assignees"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(ts.Assignments) < spec.AssignmentsMin {
		t.Fatalf("assignments %d < %d", len(ts.Assignments), spec.AssignmentsMin)
	}
	if ts.EstimatedTotal != spec.EstimatedTotal {
		t.Fatalf("estimated_total %d want %d", ts.EstimatedTotal, spec.EstimatedTotal)
	}
	if diff := cmpSortedSlicesExact(ts.RecommendedOrder, spec.RecommendedOrder); diff != "" {
		t.Fatalf("recommended_order: %s", diff)
	}
	got := make(map[string]string, len(ts.Assignments))
	for _, a := range ts.Assignments {
		got[a.Challenge] = a.Assignee
	}
	for challenge, wantAssignee := range spec.ChallengeAssignees {
		if got[challenge] != wantAssignee {
			t.Fatalf("assignee for %q got %q want %q", challenge, got[challenge], wantAssignee)
		}
	}
}

func TestGolden_AutoSolveSteps(t *testing.T) {
	mgr := NewManager()
	auto := &Automator{Manager: mgr}
	ch := Challenge{Name: "Login Portal", Category: "web", Description: "A basic login page", Difficulty: "medium", Points: 100}
	_ = ch.Validate(false)

	res := auto.AutoSolve(context.Background(), "golden-test", ch, AutoSolveOptions{ExecuteTools: false, MaxSteps: 8})

	b, err := os.ReadFile(filepath.Join(ctfGoldenDir(t), "auto_solve_steps.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		ChallengeID         string   `json:"challenge_id"`
		Status              string   `json:"status"`
		AutomatedStepsMin   int      `json:"automated_steps_min"`
		Actions             []string `json:"actions"`
		Confidence          float64  `json:"confidence"`
		PlannedOutputPrefix string   `json:"planned_output_prefix"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if res.ChallengeID != spec.ChallengeID {
		t.Fatalf("challenge_id %q want %q", res.ChallengeID, spec.ChallengeID)
	}
	if res.Status != spec.Status {
		t.Fatalf("status %q want %q", res.Status, spec.Status)
	}
	if len(res.AutomatedSteps) < spec.AutomatedStepsMin {
		t.Fatalf("steps %d < %d", len(res.AutomatedSteps), spec.AutomatedStepsMin)
	}
	if math.Abs(res.Confidence-spec.Confidence) > 1e-9 {
		t.Fatalf("confidence %v want %v", res.Confidence, spec.Confidence)
	}
	for i, wantAction := range spec.Actions {
		if i >= len(res.AutomatedSteps) {
			t.Fatalf("missing step[%d]", i)
		}
		got := res.AutomatedSteps[i]
		if got.Action != wantAction {
			t.Fatalf("step[%d] action %q want %q", i, got.Action, wantAction)
		}
		if !strings.HasPrefix(got.Output, spec.PlannedOutputPrefix) {
			t.Fatalf("step[%d] output prefix got %q", i, got.Output)
		}
		tools := sortedCopy(got.ToolsUsed)
		for _, tool := range tools {
			if tool == "manual" {
				t.Fatalf("step[%d] tools_used should omit manual, got %v", i, got.ToolsUsed)
			}
		}
	}
}

func TestGolden_CryptoSolver(t *testing.T) {
	out := AnalyzeCrypto("deadbeefcafebabe", "unknown", "", "", "")

	b, err := os.ReadFile(filepath.Join(ctfGoldenDir(t), "crypto_solver.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		CipherType             string   `json:"cipher_type"`
		AnalysisResultsSorted  []string `json:"analysis_results_sorted"`
		RecommendedToolsSorted []string `json:"recommended_tools_sorted"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if out.CipherType != spec.CipherType {
		t.Fatalf("cipher_type %q want %q", out.CipherType, spec.CipherType)
	}
	if diff := cmpSortedSlicesExact(sortedCopy(out.AnalysisResults), sortedCopy(spec.AnalysisResultsSorted)); diff != "" {
		t.Fatalf("analysis_results: %s", diff)
	}
	if diff := cmpSortedSlicesExact(sortedCopy(out.RecommendedTools), sortedCopy(spec.RecommendedToolsSorted)); diff != "" {
		t.Fatalf("recommended_tools: %s", diff)
	}
}

func TestGolden_ForensicsAnalyzer(t *testing.T) {
	opts := ForensicsOptions{AnalysisType: "quick"}
	out := AnalyzeForensics(context.Background(), "", "", opts, nil, nil)

	b, err := os.ReadFile(filepath.Join(ctfGoldenDir(t), "forensics_analyzer.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		AnalysisType            string   `json:"analysis_type"`
		RecommendedToolsSorted  []string `json:"recommended_tools_sorted"`
		HiddenDataLen           int      `json:"hidden_data_len"`
		SteganographyResultsLen int      `json:"steganography_results_len"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if out.AnalysisType != spec.AnalysisType {
		t.Fatalf("analysis_type %q want %q", out.AnalysisType, spec.AnalysisType)
	}
	if diff := cmpSortedSlicesExact(sortedCopy(out.RecommendedTools), sortedCopy(spec.RecommendedToolsSorted)); diff != "" {
		t.Fatalf("recommended_tools: %s", diff)
	}
	if len(out.HiddenData) != spec.HiddenDataLen {
		t.Fatalf("hidden_data len %d want %d", len(out.HiddenData), spec.HiddenDataLen)
	}
	if len(out.SteganographyResults) != spec.SteganographyResultsLen {
		t.Fatalf("steganography_results len %d want %d", len(out.SteganographyResults), spec.SteganographyResultsLen)
	}
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	slices.Sort(out)
	return out
}

func cmpSortedSlicesExact(a, b []string) string {
	if len(a) != len(b) {
		return fmt.Sprintf("length mismatch %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			return fmt.Sprintf("mismatch at %d: %q vs %q", i, a[i], b[i])
		}
	}
	return ""
}
