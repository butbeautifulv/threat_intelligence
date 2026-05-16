package contract

// ToolRunRequest is the JSON body for POST /api/tools/{name}.
type ToolRunRequest struct {
	Target         string            `json:"target,omitempty"`
	AdditionalArgs string            `json:"additional_args,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty"`
}

// ToolRunResponse is returned after a tool execution.
type ToolRunResponse struct {
	Success   bool   `json:"success"`
	Tool      string `json:"tool"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
	JobID     string `json:"job_id,omitempty"`
}

// AnalyzeTargetRequest mirrors HexStrike intelligence API.
type AnalyzeTargetRequest struct {
	Target       string `json:"target"`
	AnalysisType string `json:"analysis_type,omitempty"`
}

// AnalyzeTargetResponse is a simplified target profile.
type AnalyzeTargetResponse struct {
	Target      string         `json:"target"`
	TargetType  string         `json:"target_type"`
	Technologies []string      `json:"technologies,omitempty"`
	RiskLevel   string         `json:"risk_level"`
	Confidence  float64        `json:"confidence"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
