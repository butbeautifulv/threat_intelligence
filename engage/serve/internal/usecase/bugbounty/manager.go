package bugbounty

// PhaseTool is one tool invocation in a workflow phase.
type PhaseTool struct {
	Tool   string            `json:"tool"`
	Params map[string]string `json:"params,omitempty"`
}

// Phase is a named step in a bug bounty workflow.
type Phase struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Tools           []PhaseTool `json:"tools"`
	ExpectedOutputs []string    `json:"expected_outputs,omitempty"`
	EstimatedTime   int         `json:"estimated_time"`
}

// ReconWorkflow is the reconnaissance workflow response shape.
type ReconWorkflow struct {
	Target        string  `json:"target"`
	Phases        []Phase `json:"phases"`
	EstimatedTime int     `json:"estimated_time"`
	ToolsCount    int     `json:"tools_count"`
}

// VulnTest is one prioritized vulnerability test plan.
type VulnTest struct {
	VulnerabilityType string         `json:"vulnerability_type"`
	Priority          int            `json:"priority"`
	Tools             []string       `json:"tools"`
	PayloadType       string         `json:"payload_type"`
	TestScenarios     []TestScenario `json:"test_scenarios"`
	EstimatedTime     int            `json:"estimated_time"`
}

// VulnHuntWorkflow is the vulnerability hunting plan.
type VulnHuntWorkflow struct {
	Target              string     `json:"target"`
	VulnerabilityTests  []VulnTest `json:"vulnerability_tests"`
	EstimatedTime       int        `json:"estimated_time"`
	PriorityScore       int        `json:"priority_score"`
}

// BusinessLogicTest is one test entry.
type BusinessLogicTest struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Tool   string `json:"tool,omitempty"`
}

// BusinessLogicCategory groups business logic tests.
type BusinessLogicCategory struct {
	Category string              `json:"category"`
	Tests    []BusinessLogicTest `json:"tests"`
}

// BusinessLogicWorkflow is the business logic testing plan.
type BusinessLogicWorkflow struct {
	Target                 string                  `json:"target"`
	BusinessLogicTests     []BusinessLogicCategory `json:"business_logic_tests"`
	EstimatedTime          int                     `json:"estimated_time"`
	ManualTestingRequired  bool                    `json:"manual_testing_required"`
}

// OSINTPhase is one OSINT gathering phase.
type OSINTPhase struct {
	Name  string      `json:"name"`
	Tools []PhaseTool `json:"tools"`
}

// OSINTWorkflow is the OSINT plan.
type OSINTWorkflow struct {
	Target            string       `json:"target"`
	OSINTPhases       []OSINTPhase `json:"osint_phases"`
	EstimatedTime     int          `json:"estimated_time"`
	IntelligenceTypes []string     `json:"intelligence_types"`
}

// FileUploadPhase is one file upload test phase.
type FileUploadPhase struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Tools              []string `json:"tools,omitempty"`
	TestFiles          []string `json:"test_files,omitempty"`
	ExpectedFindings   []string `json:"expected_findings,omitempty"`
}

// FileUploadWorkflow is the file upload testing plan.
type FileUploadWorkflow struct {
	Target     string            `json:"target"`
	TestPhases []FileUploadPhase `json:"test_phases"`
}

// ComprehensiveAssessment combines sub-workflows.
type ComprehensiveAssessment struct {
	Target                string                 `json:"target"`
	Reconnaissance        ReconWorkflow          `json:"reconnaissance"`
	VulnerabilityHunting  VulnHuntWorkflow       `json:"vulnerability_hunting"`
	OSINT                 *OSINTWorkflow         `json:"osint,omitempty"`
	BusinessLogic         *BusinessLogicWorkflow `json:"business_logic,omitempty"`
	Summary               map[string]any         `json:"summary"`
}

// Manager builds phased bug bounty workflows.
type Manager struct{}

// NewManager returns a workflow manager.
func NewManager() *Manager {
	return &Manager{}
}

// CreateReconnaissance builds a 4-phase recon workflow.
func (m *Manager) CreateReconnaissance(t Target) ReconWorkflow {
	domain := t.Domain
	wf := ReconWorkflow{Target: domain, Phases: []Phase{
		{
			Name: "subdomain_discovery", Description: "Comprehensive subdomain enumeration",
			Tools: []PhaseTool{
				{Tool: "amass", Params: map[string]string{"mode": "enum"}},
				{Tool: "subfinder", Params: map[string]string{"additional_args": "-silent"}},
			},
			ExpectedOutputs: []string{"subdomains.txt"},
			EstimatedTime:   300,
		},
		{
			Name: "http_service_discovery", Description: "Identify live HTTP services",
			Tools: []PhaseTool{
				{Tool: "httpx", Params: map[string]string{"additional_args": "-silent -tech-detect -status-code"}},
				{Tool: "nuclei", Params: map[string]string{"severity": "info"}},
			},
			ExpectedOutputs: []string{"live_hosts.txt", "technologies.json"},
			EstimatedTime:   180,
		},
		{
			Name: "content_discovery", Description: "Discover hidden content and endpoints",
			Tools: []PhaseTool{
				{Tool: "katana", Params: map[string]string{"additional_args": "-silent"}},
				{Tool: "gau", Params: map[string]string{}},
				{Tool: "gobuster", Params: map[string]string{"mode": "dir"}},
			},
			ExpectedOutputs: []string{"endpoints.txt", "js_files.txt"},
			EstimatedTime:   600,
		},
		{
			Name: "parameter_discovery", Description: "Discover hidden parameters",
			Tools: []PhaseTool{
				{Tool: "paramspider", Params: map[string]string{}},
				{Tool: "arjun", Params: map[string]string{"method": "GET,POST"}},
			},
			ExpectedOutputs: []string{"parameters.txt"},
			EstimatedTime:   240,
		},
	}}
	return finalizeRecon(wf)
}

func finalizeRecon(wf ReconWorkflow) ReconWorkflow {
	for _, p := range wf.Phases {
		wf.EstimatedTime += p.EstimatedTime
		wf.ToolsCount += len(p.Tools)
	}
	return wf
}

// CreateVulnHunt builds vulnerability tests by priority.
func (m *Manager) CreateVulnHunt(t Target) VulnHuntWorkflow {
	wf := VulnHuntWorkflow{Target: t.Domain}
	for _, vulnType := range SortedPriorityVulns(t.PriorityVulns) {
		prof, ok := HighImpactVulns[vulnType]
		if !ok {
			continue
		}
		vt := VulnTest{
			VulnerabilityType: vulnType,
			Priority:          prof.Priority,
			Tools:             append([]string(nil), prof.Tools...),
			PayloadType:       prof.PayloadType,
			TestScenarios:     testScenariosFor(vulnType),
			EstimatedTime:     prof.Priority * 30,
		}
		wf.VulnerabilityTests = append(wf.VulnerabilityTests, vt)
		wf.EstimatedTime += vt.EstimatedTime
		wf.PriorityScore += prof.Priority
	}
	return wf
}

// CreateBusinessLogic builds business logic test categories.
func (m *Manager) CreateBusinessLogic(t Target) BusinessLogicWorkflow {
	return BusinessLogicWorkflow{
		Target: t.Domain,
		BusinessLogicTests: []BusinessLogicCategory{
			{
				Category: "Authentication Bypass",
				Tests: []BusinessLogicTest{
					{Name: "Password Reset Token Reuse", Method: "manual"},
					{Name: "JWT Algorithm Confusion", Method: "automated", Tool: "jwt_tool"},
					{Name: "Session Fixation", Method: "manual"},
				},
			},
			{
				Category: "Authorization Flaws",
				Tests: []BusinessLogicTest{
					{Name: "Horizontal Privilege Escalation", Method: "automated", Tool: "arjun"},
					{Name: "Vertical Privilege Escalation", Method: "manual"},
				},
			},
			{
				Category: "Business Process Manipulation",
				Tests: []BusinessLogicTest{
					{Name: "Race Conditions", Method: "manual"},
					{Name: "Price Manipulation", Method: "manual"},
				},
			},
		},
		EstimatedTime:         480,
		ManualTestingRequired: true,
	}
}

// CreateOSINT builds OSINT phases.
func (m *Manager) CreateOSINT(t Target) OSINTWorkflow {
	return OSINTWorkflow{
		Target: t.Domain,
		OSINTPhases: []OSINTPhase{
			{Name: "Domain Intelligence", Tools: []PhaseTool{{Tool: "whois"}, {Tool: "amass"}}},
			{Name: "Social Media Intelligence", Tools: []PhaseTool{{Tool: "theharvester"}}},
			{Name: "Email Intelligence", Tools: []PhaseTool{{Tool: "theharvester"}}},
			{Name: "Technology Intelligence", Tools: []PhaseTool{{Tool: "httpx"}, {Tool: "nuclei"}}},
		},
		EstimatedTime:     240,
		IntelligenceTypes: []string{"technical", "social", "business", "infrastructure"},
	}
}

// CreateFileUpload builds file upload test phases.
func (m *Manager) CreateFileUpload(targetURL string) FileUploadWorkflow {
	return FileUploadWorkflow{
		Target: targetURL,
		TestPhases: []FileUploadPhase{
			{Name: "reconnaissance", Description: "Identify upload endpoints", Tools: []string{"katana", "gau", "paramspider"}, ExpectedFindings: []string{"upload_forms"}},
			{Name: "baseline_testing", Description: "Test legitimate file uploads", TestFiles: []string{"image.jpg", "document.pdf"}},
			{Name: "malicious_upload", Description: "Test malicious extensions", TestFiles: []string{"shell.php", "shell.php.txt"}},
			{Name: "bypass_techniques", Description: "Extension and content-type bypass", TestFiles: []string{"shell.PhP", "polyglot.jpg"}},
		},
	}
}

// CreateComprehensive combines sub-workflows.
func (m *Manager) CreateComprehensive(t Target) ComprehensiveAssessment {
	recon := m.CreateReconnaissance(t)
	vuln := m.CreateVulnHunt(t)
	assessment := ComprehensiveAssessment{
		Target:               t.Domain,
		Reconnaissance:       recon,
		VulnerabilityHunting: vuln,
	}
	totalTime := recon.EstimatedTime + vuln.EstimatedTime
	totalTools := recon.ToolsCount
	workflowCount := 2

	if t.IncludeOSINT {
		osint := m.CreateOSINT(t)
		assessment.OSINT = &osint
		totalTime += osint.EstimatedTime
		workflowCount++
	}
	if t.IncludeBusiness {
		bl := m.CreateBusinessLogic(t)
		assessment.BusinessLogic = &bl
		totalTime += bl.EstimatedTime
		workflowCount++
	}
	assessment.Summary = map[string]any{
		"total_estimated_time": totalTime,
		"total_tools":          totalTools,
		"workflow_count":       workflowCount,
		"priority_score":       vuln.PriorityScore,
	}
	return assessment
}

// PhasesFromRecon returns phases for execution.
func PhasesFromRecon(wf ReconWorkflow) []Phase {
	return wf.Phases
}

// ToolIDsFromPhase extracts short tool ids from a phase.
func ToolIDsFromPhase(p Phase) []string {
	out := make([]string, 0, len(p.Tools))
	for _, t := range p.Tools {
		out = append(out, t.Tool)
	}
	return out
}
