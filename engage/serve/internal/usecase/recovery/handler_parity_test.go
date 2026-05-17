package recovery

import (
	"strings"
	"testing"
)

// Legacy classify patterns from hexstrike_server.py IntelligentErrorHandler._initialize_error_patterns.
func TestClassify_hexstrikePatterns(t *testing.T) {
	h := Default()
	cases := []struct {
		msg  string
		want ErrorType
	}{
		{"connection timeout while scanning", TypeTimeout},
		{"operation timed out", TypeTimeout},
		{"permission denied: /usr/bin/nmap", TypePermissionDenied},
		{"sudo required for raw sockets", TypePermissionDenied},
		{"network unreachable", TypeNetworkUnreachable},
		{"connection refused by host", TypeNetworkUnreachable},
		{"rate limit exceeded", TypeRateLimited},
		{"HTTP 429 Too Many Requests", TypeRateLimited},
		{"command not found: feroxbuster", TypeToolNotFound},
		{"executable not found", TypeToolNotFound},
		{"invalid argument: unknown flag", TypeInvalidParams},
		{"syntax error near unexpected token", TypeInvalidParams},
		{"out of memory", TypeResourceExhausted},
		{"too many open files", TypeResourceExhausted},
		{"authentication failed for user", TypeAuthenticationFailed},
		{"invalid token expired", TypeAuthenticationFailed},
		{"target unreachable", TypeTargetUnreachable},
		{"dns resolution failed", TypeTargetUnreachable},
		{"json decode error", TypeParsing},
		{"malformed response body", TypeParsing},
		{"unexpected internal failure", TypeUnknown},
	}
	for _, tc := range cases {
		t.Run(string(tc.want), func(t *testing.T) {
			if got := h.Classify(tc.msg); got != tc.want {
				t.Fatalf("Classify(%q) = %q, want %q", tc.msg, got, tc.want)
			}
		})
	}
}

func TestRecoveryStrategies_primaryAction(t *testing.T) {
	h := Default()
	cases := []struct {
		errType ErrorType
		want    RecoveryAction
	}{
		{TypeTimeout, ActionRetryWithBackoff},
		{TypePermissionDenied, ActionEscalateToHuman},
		{TypeNetworkUnreachable, ActionRetryWithBackoff},
		{TypeRateLimited, ActionRetryWithBackoff},
		{TypeToolNotFound, ActionSwitchToAlternativeTool},
		{TypeInvalidParams, ActionAdjustParameters},
		{TypeResourceExhausted, ActionRetryWithReducedScope},
		{TypeAuthenticationFailed, ActionEscalateToHuman},
		{TypeTargetUnreachable, ActionRetryWithBackoff},
		{TypeParsing, ActionAdjustParameters},
		{TypeUnknown, ActionRetryWithBackoff},
	}
	for _, tc := range cases {
		t.Run(string(tc.errType), func(t *testing.T) {
			if got := h.PrimaryAction(tc.errType); got != tc.want {
				t.Fatalf("PrimaryAction(%s) = %q, want %q", tc.errType, got, tc.want)
			}
			strategies := h.RecoveryStrategies(tc.errType)
			if len(strategies) == 0 {
				t.Fatal("expected at least one strategy")
			}
			if strategies[0].Action != tc.want {
				t.Fatalf("first strategy action = %q, want %q", strategies[0].Action, tc.want)
			}
		})
	}
}

func TestAdjustParams_hexstrikeTools(t *testing.T) {
	h := Default()
	cases := []struct {
		tool    string
		errType ErrorType
		in      map[string]string
		check   func(t *testing.T, out map[string]string)
	}{
		{
			tool: "nmap", errType: TypeTimeout, in: map[string]string{"additional_args": "-sV"},
			check: func(t *testing.T, out map[string]string) {
				if !strings.Contains(out["additional_args"], "-T2") {
					t.Fatalf("nmap timeout: additional_args=%q", out["additional_args"])
				}
			},
		},
		{
			tool: "gobuster", errType: TypeRateLimited, in: map[string]string{},
			check: func(t *testing.T, out map[string]string) {
				if out["threads"] != "5" || out["delay"] != "1" {
					t.Fatalf("gobuster rate limit: %+v", out)
				}
			},
		},
		{
			tool: "nuclei_scan", errType: TypeResourceExhausted, in: map[string]string{},
			check: func(t *testing.T, out map[string]string) {
				if out["threads"] != "5" {
					t.Fatalf("nuclei resource: %+v", out)
				}
			},
		},
		{
			tool: "ffuf", errType: TypeTimeout, in: map[string]string{},
			check: func(t *testing.T, out map[string]string) {
				if out["threads"] != "10" || out["timeout"] != "30" {
					t.Fatalf("ffuf timeout: %+v", out)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.tool+"_"+string(tc.errType), func(t *testing.T) {
			out := h.AdjustParams(tc.tool, tc.errType, tc.in)
			tc.check(t, out)
		})
	}
}

func TestParseErrorType_legacyAliases(t *testing.T) {
	cases := []struct {
		in   string
		want ErrorType
	}{
		{"rate_limited", TypeRateLimited},
		{"rate_limit", TypeRateLimited},
		{"tool_not_found", TypeToolNotFound},
		{"not_found", TypeToolNotFound},
		{"permission_denied", TypePermissionDenied},
		{"permission", TypePermissionDenied},
	}
	for _, tc := range cases {
		got, ok := ParseErrorType(tc.in)
		if !ok || got != tc.want {
			t.Fatalf("ParseErrorType(%q) = %q, %v; want %q", tc.in, got, ok, tc.want)
		}
	}
}
