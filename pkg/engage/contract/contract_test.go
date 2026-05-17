package contract

import (
	"encoding/json"
	"testing"
)

func TestToolRunRequest_JSONRoundTrip(t *testing.T) {
	in := ToolRunRequest{
		Target:         "https://example.com",
		AdditionalArgs: "-silent",
		Parameters:     map[string]string{"use_cache": "false"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out ToolRunRequest
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Target != in.Target || out.Parameters["use_cache"] != "false" {
		t.Fatalf("got %+v", out)
	}
}

func TestToolRunResponse_JSONFields(t *testing.T) {
	raw := `{"success":true,"tool":"nmap_scan","output":"ok","job_id":"j1"}`
	var out ToolRunResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Success || out.Tool != "nmap_scan" || out.JobID != "j1" {
		t.Fatalf("got %+v", out)
	}
}

func TestAnalyzeTargetRequestResponse_roundTrip(t *testing.T) {
	req := AnalyzeTargetRequest{Target: "10.0.0.1", AnalysisType: "quick"}
	rb, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var req2 AnalyzeTargetRequest
	if err := json.Unmarshal(rb, &req2); err != nil {
		t.Fatal(err)
	}
	if req2.Target != req.Target {
		t.Fatalf("req %+v", req2)
	}

	resp := AnalyzeTargetResponse{
		Target:     "10.0.0.1",
		TargetType: "ip",
		RiskLevel:  "low",
		Confidence: 0.9,
		Metadata:   map[string]any{"ports": []any{float64(443)}},
	}
	sb, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var resp2 AnalyzeTargetResponse
	if err := json.Unmarshal(sb, &resp2); err != nil {
		t.Fatal(err)
	}
	if resp2.Confidence != resp.Confidence || resp2.RiskLevel != resp.RiskLevel {
		t.Fatalf("resp %+v", resp2)
	}
}
