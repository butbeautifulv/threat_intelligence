package bugbounty

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func goldenDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "golden")
}

func TestGolden_ReconWorkflow(t *testing.T) {
	m := NewManager()
	wf := m.CreateReconnaissance(Target{Domain: "example.com"})
	names := phaseNames(wf.Phases)

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "recon_workflow.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		PhasesMin        int      `json:"phases_min"`
		ToolsCountMin    int      `json:"tools_count_min"`
		EstimatedTimeMin int      `json:"estimated_time_min"`
		PhaseNames       []string `json:"phase_names"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.Phases) < spec.PhasesMin {
		t.Fatalf("phases %d < %d", len(wf.Phases), spec.PhasesMin)
	}
	if wf.ToolsCount < spec.ToolsCountMin {
		t.Fatalf("tools_count %d < %d", wf.ToolsCount, spec.ToolsCountMin)
	}
	if wf.EstimatedTime < spec.EstimatedTimeMin {
		t.Fatalf("estimated_time %d < %d", wf.EstimatedTime, spec.EstimatedTimeMin)
	}
	assertPhaseNames(t, names, spec.PhaseNames)
}

func TestGolden_VulnHuntWorkflow(t *testing.T) {
	m := NewManager()
	wf := m.CreateVulnHunt(Target{
		Domain:        "example.com",
		PriorityVulns: []string{"rce", "sqli", "xss"},
	})

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "vuln_hunt_workflow.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		VulnerabilityTestsMin int      `json:"vulnerability_tests_min"`
		VulnerabilityTypes    []string `json:"vulnerability_types"`
		EstimatedTimeMin      int      `json:"estimated_time_min"`
		PriorityScoreMin      int      `json:"priority_score_min"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.VulnerabilityTests) < spec.VulnerabilityTestsMin {
		t.Fatalf("tests %d < %d", len(wf.VulnerabilityTests), spec.VulnerabilityTestsMin)
	}
	if wf.EstimatedTime < spec.EstimatedTimeMin {
		t.Fatalf("estimated_time %d < %d", wf.EstimatedTime, spec.EstimatedTimeMin)
	}
	if wf.PriorityScore < spec.PriorityScoreMin {
		t.Fatalf("priority_score %d < %d", wf.PriorityScore, spec.PriorityScoreMin)
	}
	gotTypes := make([]string, len(wf.VulnerabilityTests))
	for i, vt := range wf.VulnerabilityTests {
		gotTypes[i] = vt.VulnerabilityType
	}
	if diff := cmpSortedSlices(gotTypes, spec.VulnerabilityTypes); diff != "" {
		t.Fatal(diff)
	}
}

func TestGolden_BusinessLogicWorkflow(t *testing.T) {
	m := NewManager()
	wf := m.CreateBusinessLogic(Target{Domain: "example.com"})

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "business_logic_workflow.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		CategoriesMin         int      `json:"categories_min"`
		TestsMin              int      `json:"tests_min"`
		EstimatedTimeMin      int      `json:"estimated_time_min"`
		ManualTestingRequired bool     `json:"manual_testing_required"`
		CategoryNames         []string `json:"category_names"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.BusinessLogicTests) < spec.CategoriesMin {
		t.Fatalf("categories %d < %d", len(wf.BusinessLogicTests), spec.CategoriesMin)
	}
	totalTests := 0
	names := make([]string, len(wf.BusinessLogicTests))
	for i, c := range wf.BusinessLogicTests {
		names[i] = c.Category
		totalTests += len(c.Tests)
	}
	if totalTests < spec.TestsMin {
		t.Fatalf("tests %d < %d", totalTests, spec.TestsMin)
	}
	if wf.EstimatedTime < spec.EstimatedTimeMin {
		t.Fatalf("estimated_time %d < %d", wf.EstimatedTime, spec.EstimatedTimeMin)
	}
	if wf.ManualTestingRequired != spec.ManualTestingRequired {
		t.Fatalf("manual_testing_required got %v want %v", wf.ManualTestingRequired, spec.ManualTestingRequired)
	}
	assertPhaseNames(t, names, spec.CategoryNames)
}

func TestGolden_OSINTWorkflow(t *testing.T) {
	m := NewManager()
	wf := m.CreateOSINT(Target{Domain: "example.com"})

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "osint_workflow.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		PhasesMin               int      `json:"phases_min"`
		ToolsCountMin           int      `json:"tools_count_min"`
		EstimatedTimeMin        int      `json:"estimated_time_min"`
		PhaseNames              []string `json:"phase_names"`
		IntelligenceTypesSorted []string `json:"intelligence_types_sorted"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.OSINTPhases) < spec.PhasesMin {
		t.Fatalf("phases %d < %d", len(wf.OSINTPhases), spec.PhasesMin)
	}
	tools := 0
	names := make([]string, len(wf.OSINTPhases))
	for i, p := range wf.OSINTPhases {
		names[i] = p.Name
		tools += len(p.Tools)
	}
	if tools < spec.ToolsCountMin {
		t.Fatalf("tools_count %d < %d", tools, spec.ToolsCountMin)
	}
	if wf.EstimatedTime < spec.EstimatedTimeMin {
		t.Fatalf("estimated_time %d < %d", wf.EstimatedTime, spec.EstimatedTimeMin)
	}
	assertPhaseNames(t, names, spec.PhaseNames)

	gotIntel := append([]string(nil), wf.IntelligenceTypes...)
	slices.Sort(gotIntel)
	if diff := cmpSortedSlices(gotIntel, spec.IntelligenceTypesSorted); diff != "" {
		t.Fatal(diff)
	}
}

func TestGolden_FileUploadWorkflow(t *testing.T) {
	m := NewManager()
	wf := m.CreateFileUpload("https://example.com/upload")

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "file_upload_workflow.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		PhasesMin  int      `json:"phases_min"`
		PhaseNames []string `json:"phase_names"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.TestPhases) < spec.PhasesMin {
		t.Fatalf("phases %d < %d", len(wf.TestPhases), spec.PhasesMin)
	}
	names := make([]string, len(wf.TestPhases))
	for i, p := range wf.TestPhases {
		names[i] = p.Name
	}
	assertPhaseNames(t, names, spec.PhaseNames)
}

func TestGolden_FileUploadTools(t *testing.T) {
	m := NewManager()
	wf := m.CreateFileUpload("https://example.com/upload")

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "file_upload_tools.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		PhasesMin      int      `json:"phases_min"`
		PhaseNames     []string `json:"phase_names"`
		ToolsCountMin  int      `json:"tools_count_min"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.TestPhases) < spec.PhasesMin {
		t.Fatalf("phases %d < %d", len(wf.TestPhases), spec.PhasesMin)
	}
	names := make([]string, len(wf.TestPhases))
	tools := 0
	for i, p := range wf.TestPhases {
		names[i] = p.Name
		tools += len(p.Tools)
	}
	assertPhaseNames(t, names, spec.PhaseNames)
	if tools < spec.ToolsCountMin {
		t.Fatalf("tools_count %d < %d", tools, spec.ToolsCountMin)
	}
}

func TestGolden_VulnHuntDefaultPriority(t *testing.T) {
	m := NewManager()
	wf := m.CreateVulnHunt(Target{
		Domain:        "example.com",
		PriorityVulns: []string{"rce", "sqli", "xss", "idor", "ssrf"},
	})

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "vuln_hunt_default_priority.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		VulnerabilityTestsMin int      `json:"vulnerability_tests_min"`
		VulnerabilityTypes    []string `json:"vulnerability_types"`
		EstimatedTimeMin      int      `json:"estimated_time_min"`
		PriorityScoreMin      int      `json:"priority_score_min"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.VulnerabilityTests) < spec.VulnerabilityTestsMin {
		t.Fatalf("tests %d < %d", len(wf.VulnerabilityTests), spec.VulnerabilityTestsMin)
	}
	if wf.EstimatedTime < spec.EstimatedTimeMin {
		t.Fatalf("estimated_time %d < %d", wf.EstimatedTime, spec.EstimatedTimeMin)
	}
	if wf.PriorityScore < spec.PriorityScoreMin {
		t.Fatalf("priority_score %d < %d", wf.PriorityScore, spec.PriorityScoreMin)
	}
	gotTypes := make([]string, len(wf.VulnerabilityTests))
	for i, vt := range wf.VulnerabilityTests {
		gotTypes[i] = vt.VulnerabilityType
	}
	if diff := cmpSortedSlices(gotTypes, spec.VulnerabilityTypes); diff != "" {
		t.Fatal(diff)
	}
}

func TestGolden_VulnHuntScenarios(t *testing.T) {
	m := NewManager()
	wf := m.CreateVulnHunt(Target{
		Domain:        "example.com",
		PriorityVulns: []string{"rce", "sqli", "xss", "idor", "ssrf"},
	})

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "vuln_hunt_scenarios.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		ScenarioCounts map[string]int `json:"scenario_counts"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	got := make(map[string]int, len(wf.VulnerabilityTests))
	for _, vt := range wf.VulnerabilityTests {
		got[vt.VulnerabilityType] = len(vt.TestScenarios)
	}
	for vulnType, wantCount := range spec.ScenarioCounts {
		if got[vulnType] != wantCount {
			t.Fatalf("%s scenarios %d want %d", vulnType, got[vulnType], wantCount)
		}
	}
}

func TestGolden_ComprehensiveMinimal(t *testing.T) {
	m := NewManager()
	a := m.CreateComprehensive(Target{
		Domain:          "example.com",
		PriorityVulns:   []string{"rce", "sqli"},
		IncludeOSINT:    false,
		IncludeBusiness: false,
	})

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "comprehensive_minimal.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		WorkflowCount         int  `json:"workflow_count"`
		TotalEstimatedTimeMin int  `json:"total_estimated_time_min"`
		TotalToolsMin         int  `json:"total_tools_min"`
		PriorityScoreMin      int  `json:"priority_score_min"`
		IncludesOSINT         bool `json:"includes_osint"`
		IncludesBusinessLogic bool `json:"includes_business_logic"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if a.OSINT != nil {
		t.Fatal("unexpected osint")
	}
	if a.BusinessLogic != nil {
		t.Fatal("unexpected business logic")
	}
	if got := a.Summary["workflow_count"].(int); got != spec.WorkflowCount {
		t.Fatalf("workflow_count %d want %d", got, spec.WorkflowCount)
	}
	if got := a.Summary["total_estimated_time"].(int); got < spec.TotalEstimatedTimeMin {
		t.Fatalf("total_estimated_time %d < %d", got, spec.TotalEstimatedTimeMin)
	}
	if got := a.Summary["total_tools"].(int); got < spec.TotalToolsMin {
		t.Fatalf("total_tools %d < %d", got, spec.TotalToolsMin)
	}
	if got := a.Summary["priority_score"].(int); got < spec.PriorityScoreMin {
		t.Fatalf("priority_score %d < %d", got, spec.PriorityScoreMin)
	}
}

func TestGolden_ComprehensiveAssessment(t *testing.T) {
	m := NewManager()
	target := Target{
		Domain:          "example.com",
		PriorityVulns:   []string{"rce", "sqli", "xss"},
		IncludeOSINT:    true,
		IncludeBusiness: true,
	}
	a := m.CreateComprehensive(target)

	b, err := os.ReadFile(filepath.Join(goldenDir(t), "comprehensive_assessment.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		WorkflowCount         int  `json:"workflow_count"`
		TotalEstimatedTimeMin int  `json:"total_estimated_time_min"`
		TotalToolsMin         int  `json:"total_tools_min"`
		PriorityScoreMin      int  `json:"priority_score_min"`
		IncludesOSINT         bool `json:"includes_osint"`
		IncludesBusinessLogic bool `json:"includes_business_logic"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if spec.IncludesOSINT && a.OSINT == nil {
		t.Fatal("missing osint")
	}
	if spec.IncludesBusinessLogic && a.BusinessLogic == nil {
		t.Fatal("missing business logic")
	}
	if got := a.Summary["workflow_count"].(int); got != spec.WorkflowCount {
		t.Fatalf("workflow_count %d want %d", got, spec.WorkflowCount)
	}
	if got := a.Summary["total_estimated_time"].(int); got < spec.TotalEstimatedTimeMin {
		t.Fatalf("total_estimated_time %d < %d", got, spec.TotalEstimatedTimeMin)
	}
	if got := a.Summary["total_tools"].(int); got < spec.TotalToolsMin {
		t.Fatalf("total_tools %d < %d", got, spec.TotalToolsMin)
	}
	if got := a.Summary["priority_score"].(int); got < spec.PriorityScoreMin {
		t.Fatalf("priority_score %d < %d", got, spec.PriorityScoreMin)
	}
}

func phaseNames(phases []Phase) []string {
	out := make([]string, len(phases))
	for i, p := range phases {
		out[i] = p.Name
	}
	return out
}

func assertPhaseNames(t *testing.T, got, want []string) {
	t.Helper()
	for i, w := range want {
		if i >= len(got) {
			t.Fatalf("missing phase[%d] want %q", i, w)
		}
		if got[i] != w {
			t.Fatalf("phase[%d] %q != %q", i, got[i], w)
		}
	}
}

func cmpSortedSlices(a, b []string) string {
	a = append([]string(nil), a...)
	b = append([]string(nil), b...)
	slices.Sort(a)
	slices.Sort(b)
	if len(a) != len(b) {
		return "length mismatch"
	}
	for i := range a {
		if a[i] != b[i] {
			return "slice mismatch at " + a[i] + " vs " + b[i]
		}
	}
	return ""
}
