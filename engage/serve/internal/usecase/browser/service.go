package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// InspectRequest configures browser inspection.
type InspectRequest struct {
	URL         string
	WaitTime    int
	Headless    bool
	Screenshot  bool
	ActiveTests bool
}

// InspectResult is the normalized browser inspect response.
type InspectResult struct {
	Success           bool           `json:"success"`
	PageInfo          map[string]any `json:"page_info,omitempty"`
	SecurityAnalysis  map[string]any `json:"security_analysis,omitempty"`
	Technologies      []string       `json:"technologies,omitempty"`
	Screenshot        string         `json:"screenshot,omitempty"`
	Timestamp         string         `json:"timestamp,omitempty"`
	Error             string         `json:"error,omitempty"`
}

// Service calls the Playwright browser sidecar.
type Service struct {
	BaseURL string
	Client  *http.Client
}

// NewServiceFromEnv returns a browser service when ENGAGE_BROWSER_URL is set.
func NewServiceFromEnv() *Service {
	base := strings.TrimSpace(os.Getenv("ENGAGE_BROWSER_URL"))
	if base == "" {
		return nil
	}
	return &Service{
		BaseURL: strings.TrimRight(base, "/"),
		Client:  &http.Client{Timeout: 5 * time.Minute},
	}
}

func (s *Service) Enabled() bool {
	return s != nil && s.BaseURL != ""
}

// Inspect runs POST /inspect on the sidecar.
func (s *Service) Inspect(ctx context.Context, req InspectRequest) InspectResult {
	if s == nil || !s.Enabled() {
		return InspectResult{Success: false, Error: "browser sidecar not configured (ENGAGE_BROWSER_URL)"}
	}
	url := strings.TrimSpace(req.URL)
	if url == "" {
		return InspectResult{Success: false, Error: "url required"}
	}
	payload := map[string]any{
		"url":           url,
		"target":        url,
		"wait_time":     req.WaitTime,
		"headless":      req.Headless,
		"screenshot":    req.Screenshot,
		"active_tests":  req.ActiveTests,
		"inspect":       true,
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/inspect", bytes.NewReader(body))
	if err != nil {
		return InspectResult{Success: false, Error: err.Error()}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return InspectResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	var out InspectResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return InspectResult{Success: false, Error: fmt.Sprintf("decode: %v", err)}
	}
	if resp.StatusCode >= 400 && out.Error == "" {
		out.Success = false
		out.Error = fmt.Sprintf("browser sidecar http %d", resp.StatusCode)
	}
	return out
}

// InspectFromParams maps catalog/MCP parameters to InspectRequest.
func InspectFromParams(target string, params map[string]string) InspectRequest {
	url := target
	if params != nil {
		if u := params["url"]; u != "" {
			url = u
		}
	}
	wait := 5
	if params != nil {
		if w := params["wait_time"]; w != "" {
			fmt.Sscanf(w, "%d", &wait)
		}
	}
	headless := true
	if params != nil && (params["headless"] == "false" || params["headless"] == "False") {
		headless = false
	}
	active := false
	if params != nil && (params["active_tests"] == "true" || params["active_tests"] == "True") {
		active = true
	}
	return InspectRequest{
		URL:         url,
		WaitTime:    wait,
		Headless:    headless,
		Screenshot:  true,
		ActiveTests: active,
	}
}

// ToJSON returns inspect result as JSON string for tool output.
func (r InspectResult) ToJSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}
