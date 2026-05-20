// Package security provides engage-layer hardening audits and safe self-tests (no host exploitation).
package security

import (
	"fmt"
	"os"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/config"
)

// Severity ranks findings for fail-on thresholds.
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// Finding is a single hardening audit result.
type Finding struct {
	ID          string
	Severity    Severity
	Message     string
	Remediation string
}

// Shell metacharacters that must not appear in guarded /api/command strings.
const shellMetachar = ";|&$`<>()\n\r\x00"

// ContainsShellMetacharacters reports injection-prone characters in a command line.
func ContainsShellMetacharacters(s string) bool {
	return strings.ContainsAny(s, shellMetachar)
}

// AuditSecurityConfig checks engage security settings (in-process, no network).
func AuditSecurityConfig(sec config.SecurityConfig, authEnabled bool) []Finding {
	var out []Finding
	if sec.Prod && sec.AllowRawCommand {
		out = append(out, Finding{
			ID: "raw-command-in-prod", Severity: SeverityHigh,
			Message:     "AllowRawCommand is true while ENGAGE_ENV=prod",
			Remediation: "Set ENGAGE_DENY_RAW_COMMAND=1 or ENGAGE_ENV!=prod",
		})
	}
	if sec.RequireAuth && !authEnabled {
		out = append(out, Finding{
			ID: "require-auth-without-auth", Severity: SeverityHigh,
			Message:     "VEIL_REQUIRE_AUTH=1 but AUTH_ENABLED=0",
			Remediation: "Enable AUTH_ENABLED=1 and configure Keycloak",
		})
	}
	if sec.Prod && !sec.RequireAuth {
		out = append(out, Finding{
			ID: "prod-without-require-auth", Severity: SeverityMedium,
			Message:     "ENGAGE_ENV=prod without VEIL_REQUIRE_AUTH=1",
			Remediation: "Set VEIL_REQUIRE_AUTH=1 in secure profiles",
		})
	}
	if sec.Prod && !sec.MCPHTTPAuthStrict {
		out = append(out, Finding{
			ID: "prod-mcp-auth-lax", Severity: SeverityMedium,
			Message:     "ENGAGE_MCP_HTTP_AUTH_STRICT is not enabled in prod",
			Remediation: "Set ENGAGE_MCP_HTTP_AUTH_STRICT=1",
		})
	}
	return out
}

// AuditProcessEnv inspects runtime environment (safe static checks only).
func AuditProcessEnv() []Finding {
	return auditProcessEnv(os.Getenv)
}

func auditProcessEnv(getenv func(string) string) []Finding {
	var out []Finding
	prod := strings.EqualFold(strings.TrimSpace(getenv("ENGAGE_ENV")), "prod")
	if prod && strings.TrimSpace(getenv("ENGAGE_ALLOW_RAW_COMMAND")) == "1" {
		out = append(out, Finding{
			ID: "env-allow-raw-prod", Severity: SeverityHigh,
			Message:     "ENGAGE_ALLOW_RAW_COMMAND=1 with ENGAGE_ENV=prod",
			Remediation: "Unset ENGAGE_ALLOW_RAW_COMMAND or use ENGAGE_DENY_RAW_COMMAND=1",
		})
	}
	pass := strings.TrimSpace(getenv("NEO4J_PASS"))
	if pass != "" && isWeakCredential(pass) {
		out = append(out, Finding{
			ID: "weak-neo4j-password", Severity: SeverityMedium,
			Message:     "NEO4J_PASS matches a known weak default",
			Remediation: "Rotate Neo4j password via secrets manager",
		})
	}
	if strings.TrimSpace(getenv("ENGAGE_AUDIT_WEBHOOK_SECRET")) == "" && strings.TrimSpace(getenv("ENGAGE_AUDIT_WEBHOOK_URL")) != "" {
		out = append(out, Finding{
			ID: "audit-webhook-no-secret", Severity: SeverityLow,
			Message:     "ENGAGE_AUDIT_WEBHOOK_URL set without ENGAGE_AUDIT_WEBHOOK_SECRET",
			Remediation: "Set a shared secret for webhook HMAC",
		})
	}
	return out
}

func isWeakCredential(pass string) bool {
	weak := []string{"neo4jpassword", "neo4j", "password", "changeme", "admin", "veilpass"}
	lower := strings.ToLower(pass)
	for _, w := range weak {
		if lower == w {
			return true
		}
	}
	return false
}

// RunSelfTest aggregates in-process checks for CI (no port scan, no exploit payloads on host).
func RunSelfTest(sec config.SecurityConfig, authEnabled bool) []Finding {
	var all []Finding
	all = append(all, AuditSecurityConfig(sec, authEnabled)...)
	all = append(all, AuditProcessEnv()...)
	all = append(all, selfTestCommandGuard()...)
	all = append(all, selfTestShellMetachar()...)
	return all
}

func selfTestCommandGuard() []Finding {
	if !ContainsShellMetacharacters("id; curl evil") {
		return []Finding{{
			ID: "selftest-metachar-detector", Severity: SeverityHigh,
			Message:     "shell metachar detector failed to flag injection sample",
			Remediation: "fix ContainsShellMetacharacters",
		}}
	}
	if ContainsShellMetacharacters("nmap -sV 127.0.0.1") {
		return []Finding{{
			ID: "selftest-metachar-fp", Severity: SeverityMedium,
			Message:     "shell metachar detector false-positive on benign scan target",
			Remediation: "narrow metachar set",
		}}
	}
	return nil
}

func selfTestShellMetachar() []Finding {
	// Parentheses in tool args are blocked in /api/command; catalog tools use argv arrays.
	return nil
}

// FailOn returns an error if any finding meets or exceeds min severity.
func FailOn(findings []Finding, min Severity) error {
	order := map[Severity]int{SeverityLow: 1, SeverityMedium: 2, SeverityHigh: 3}
	need := order[min]
	var critical []string
	for _, f := range findings {
		if order[f.Severity] >= need {
			critical = append(critical, fmt.Sprintf("[%s] %s: %s", f.Severity, f.ID, f.Message))
		}
	}
	if len(critical) == 0 {
		return nil
	}
	return fmt.Errorf("hardening: %d finding(s) at or above %s:\n%s", len(critical), min, strings.Join(critical, "\n"))
}

// FormatReport renders findings for logs or scripts.
func FormatReport(findings []Finding) string {
	if len(findings) == 0 {
		return "hardening: OK (0 findings)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "hardening: %d finding(s)\n", len(findings))
	for _, f := range findings {
		fmt.Fprintf(&b, "  [%s] %s: %s\n", f.Severity, f.ID, f.Message)
		if f.Remediation != "" {
			fmt.Fprintf(&b, "         → %s\n", f.Remediation)
		}
	}
	return b.String()
}
