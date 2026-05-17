package cve

import "testing"

func TestAnalyzeExploitability_log4j(t *testing.T) {
	a := AnalyzeExploitability(CVEEntry{
		CVEID:       "CVE-2021-44228",
		Description: "Apache Log4j remote code execution via JNDI",
		Severity:    "CRITICAL",
		CVSSScore:   10,
	})
	if !a.Success {
		t.Fatal("expected success")
	}
	if a.VulnerabilityType != "rce" {
		t.Fatalf("vuln type %q", a.VulnerabilityType)
	}
	if a.ExploitabilityLevel != "CRITICAL" {
		t.Fatalf("level %q", a.ExploitabilityLevel)
	}
	if a.ExploitabilityScore <= 0 {
		t.Fatalf("score %v", a.ExploitabilityScore)
	}
}

func TestClassifyVuln(t *testing.T) {
	cases := map[string]string{
		"SQL injection in login":        "sql_injection",
		"stored cross-site scripting":   "xss",
		"XML external entity expansion":   "xxe",
		"unsafe deserialization":        "deserialization",
		"remote code execution":         "rce",
		"authentication bypass in auth": "authentication_bypass",
		"stack buffer overflow":         "buffer_overflow",
	}
	for desc, want := range cases {
		if got := ClassifyVuln(desc); got != want {
			t.Fatalf("%q: got %q want %q", desc, got, want)
		}
	}
}
