package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/recovery"
)

func errorHandlingMux(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	registerErrorHandling(mux)
	return mux
}

func postErrorHandling(t *testing.T, mux *http.ServeMux, path string, body map[string]any) map[string]any {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST %s status %d body %s", path, rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestPostClassifyError_hexstrikeParity(t *testing.T) {
	mux := errorHandlingMux(t)
	cases := []struct {
		msg            string
		wantType       string
		wantRecoverable bool
		wantPrimary    string
	}{
		{"operation timed out", "timeout", true, string(recovery.ActionRetryWithBackoff)},
		{"permission denied", "permission_denied", false, string(recovery.ActionEscalateToHuman)},
		{"rate limit exceeded", "rate_limited", true, string(recovery.ActionRetryWithBackoff)},
		{"command not found", "tool_not_found", true, string(recovery.ActionSwitchToAlternativeTool)},
		{"invalid argument", "invalid_parameters", true, string(recovery.ActionAdjustParameters)},
		{"authentication failed", "authentication_failed", false, string(recovery.ActionEscalateToHuman)},
	}
	for _, tc := range cases {
		t.Run(tc.wantType, func(t *testing.T) {
			resp := postErrorHandling(t, mux, "/api/error-handling/classify-error", map[string]any{
				"error_message": tc.msg,
			})
			if resp["error_type"] != tc.wantType {
				t.Fatalf("error_type=%v want %s", resp["error_type"], tc.wantType)
			}
			if resp["recoverable"] != tc.wantRecoverable {
				t.Fatalf("recoverable=%v want %v", resp["recoverable"], tc.wantRecoverable)
			}
			if resp["primary_action"] != tc.wantPrimary {
				t.Fatalf("primary_action=%v want %s", resp["primary_action"], tc.wantPrimary)
			}
			strategies, ok := resp["recovery_strategies"].([]any)
			if !ok || len(strategies) == 0 {
				t.Fatalf("missing recovery_strategies: %v", resp["recovery_strategies"])
			}
		})
	}
}

func TestPostParameterAdjustments_hexstrikeParity(t *testing.T) {
	mux := errorHandlingMux(t)
	cases := []struct {
		name    string
		body    map[string]any
		check   func(t *testing.T, adjusted map[string]any)
	}{
		{
			name: "nmap_timeout_legacy_fields",
			body: map[string]any{
				"tool_name":   "nmap",
				"error_type":  "timeout",
				"original_params": map[string]any{"additional_args": "-sV"},
			},
			check: func(t *testing.T, adjusted map[string]any) {
				args, _ := adjusted["additional_args"].(string)
				if args == "" || !strings.Contains(args, "-T2") {
					t.Fatalf("adjusted additional_args=%q", args)
				}
			},
		},
		{
			name: "gobuster_rate_limited",
			body: map[string]any{
				"tool_name":  "gobuster",
				"error_type": "rate_limited",
			},
			check: func(t *testing.T, adjusted map[string]any) {
				if adjusted["threads"] != "5" || adjusted["delay"] != "1" {
					t.Fatalf("adjusted=%v", adjusted)
				}
			},
		},
		{
			name: "engage_alias_error_type",
			body: map[string]any{
				"tool":       "feroxbuster",
				"error_type": "rate_limit",
				"params":     map[string]any{},
			},
			check: func(t *testing.T, adjusted map[string]any) {
				if adjusted["threads"] != "5" {
					t.Fatalf("adjusted=%v", adjusted)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := postErrorHandling(t, mux, "/api/error-handling/parameter-adjustments", tc.body)
			adjusted, ok := resp["adjusted_params"].(map[string]any)
			if !ok {
				adjusted, ok = resp["params"].(map[string]any)
			}
			if !ok {
				t.Fatalf("missing adjusted_params: %v", resp)
			}
			tc.check(t, adjusted)
		})
	}
}

func TestPostExecuteWithRecovery_hexstrikeParity(t *testing.T) {
	mux := errorHandlingMux(t)
	resp := postErrorHandling(t, mux, "/api/error-handling/execute-with-recovery", map[string]any{
		"tool_name":     "nuclei_scan",
		"error_message": "command not found",
	})
	if resp["error_type"] != "tool_not_found" {
		t.Fatalf("error_type=%v", resp["error_type"])
	}
	if resp["alternative"] != "jaeles" && resp["alternative"] != "nikto" {
		t.Fatalf("alternative=%v", resp["alternative"])
	}
	if resp["primary_action"] != string(recovery.ActionSwitchToAlternativeTool) {
		t.Fatalf("primary_action=%v", resp["primary_action"])
	}
}
