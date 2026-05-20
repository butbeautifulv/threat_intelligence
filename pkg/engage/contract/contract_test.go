package contract

import (
	"encoding/json"
	"strings"
	"testing"
)

func decodeJSON[T any](t *testing.T, raw string) T {
	t.Helper()
	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("decode %T: %v", out, err)
	}
	return out
}

func TestDecode_invalidJSON_errors(t *testing.T) {
	bad := []byte(`{not json`)
	types := []struct {
		name string
		fn   func() error
	}{
		{"ToolRunRequest", func() error { var v ToolRunRequest; return json.Unmarshal(bad, &v) }},
		{"ToolRunResponse", func() error { var v ToolRunResponse; return json.Unmarshal(bad, &v) }},
		{"AnalyzeTargetRequest", func() error { var v AnalyzeTargetRequest; return json.Unmarshal(bad, &v) }},
		{"AnalyzeTargetResponse", func() error { var v AnalyzeTargetResponse; return json.Unmarshal(bad, &v) }},
	}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err == nil {
				t.Fatal("expected decode error")
			}
		})
	}
}

func TestDecode_missingRequiredFields_zeroValues(t *testing.T) {
	t.Run("ToolRunRequest", func(t *testing.T) {
		out := decodeJSON[ToolRunRequest](t, `{}`)
		if out.Target != "" || out.AdditionalArgs != "" || len(out.Parameters) != 0 {
			t.Fatalf("got %+v", out)
		}
	})
	t.Run("ToolRunResponse", func(t *testing.T) {
		out := decodeJSON[ToolRunResponse](t, `{"success":false,"tool":"nmap"}`)
		if out.Success || out.Tool != "nmap" || out.Output != "" || out.Error != "" {
			t.Fatalf("got %+v", out)
		}
	})
	t.Run("AnalyzeTargetRequest", func(t *testing.T) {
		out := decodeJSON[AnalyzeTargetRequest](t, `{}`)
		if out.Target != "" {
			t.Fatalf("missing target should be empty, got %q", out.Target)
		}
		if out.AnalysisType != "" {
			t.Fatalf("got %+v", out)
		}
	})
	t.Run("AnalyzeTargetResponse", func(t *testing.T) {
		out := decodeJSON[AnalyzeTargetResponse](t, `{}`)
		if out.Target != "" || out.TargetType != "" || out.RiskLevel != "" || out.Confidence != 0 {
			t.Fatalf("got %+v", out)
		}
	})
}

func TestDecode_allRequestResponseTypes(t *testing.T) {
	toolReq := decodeJSON[ToolRunRequest](t, `{"target":"host","additional_args":"-v","parameters":{"k":"v"}}`)
	if toolReq.Target != "host" || toolReq.Parameters["k"] != "v" {
		t.Fatalf("ToolRunRequest %+v", toolReq)
	}

	toolResp := decodeJSON[ToolRunResponse](t, `{"success":true,"tool":"nmap","output":"scan","exit_code":0,"job_id":"j-1"}`)
	if !toolResp.Success || toolResp.ExitCode != 0 || toolResp.JobID != "j-1" {
		t.Fatalf("ToolRunResponse %+v", toolResp)
	}

	analyzeReq := decodeJSON[AnalyzeTargetRequest](t, `{"target":"10.0.0.2","analysis_type":"deep"}`)
	if analyzeReq.Target != "10.0.0.2" || analyzeReq.AnalysisType != "deep" {
		t.Fatalf("AnalyzeTargetRequest %+v", analyzeReq)
	}

	analyzeResp := decodeJSON[AnalyzeTargetResponse](t, `{"target":"10.0.0.2","target_type":"ip","technologies":["nginx"],"risk_level":"medium","confidence":0.5,"metadata":{"tier":1}}`)
	if analyzeResp.TargetType != "ip" || len(analyzeResp.Technologies) != 1 {
		t.Fatalf("AnalyzeTargetResponse %+v", analyzeResp)
	}
	if analyzeResp.Metadata["tier"] != float64(1) {
		t.Fatalf("metadata %+v", analyzeResp.Metadata)
	}
}

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

func TestToolRunResponse_errorFields(t *testing.T) {
	raw := `{"success":false,"tool":"nuclei","error":"timeout","exit_code":124}`
	out := decodeJSON[ToolRunResponse](t, raw)
	if out.Success || !strings.Contains(out.Error, "timeout") || out.ExitCode != 124 {
		t.Fatalf("got %+v", out)
	}
}
