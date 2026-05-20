package decision

import "testing"

func TestDecisionEngine_CandidateTools_and_Score(t *testing.T) {
	d := DefaultDecisionEngine()
	cands := d.CandidateTools("web")
	if len(cands) < 3 {
		t.Fatalf("web candidates: %d", len(cands))
	}
	unknown := d.CandidateTools("not-a-real-type")
	if len(unknown) == 0 {
		t.Fatal("expected unknown fallback table")
	}
	if d.Score("web", "nuclei") <= d.Score("web", "missing-tool") {
		t.Fatalf("nuclei should beat default: %v vs %v", d.Score("web", "nuclei"), d.Score("web", "missing-tool"))
	}
}

func TestChainMetrics_helpers(t *testing.T) {
	if ExecutionTimeEstimate("nmap") != 120 {
		t.Fatalf("nmap estimate: %d", ExecutionTimeEstimate("nmap"))
	}
	if ExecutionTimeEstimate("unknown-tool-xyz") != 180 {
		t.Fatalf("default estimate: %d", ExecutionTimeEstimate("unknown-tool-xyz"))
	}
	out := ExpectedOutcome("nuclei")
	if out == "" || out == "Discover vulnerabilities using " {
		t.Fatalf("outcome: %q", out)
	}
	if StepSuccessProbability(0.8, 0.5) != 0.4 {
		t.Fatalf("probability: %v", StepSuccessProbability(0.8, 0.5))
	}
}

func TestSelectPatternKey_objectives(t *testing.T) {
	cases := []struct {
		tt, obj, want string
	}{
		{"api", "recon", "api_testing"},
		{"web", "recon", "web_reconnaissance"},
		{"ip", "comprehensive", "comprehensive_network_pentest"},
		{"ip", "", "network_discovery"},
		{"web", "vuln-hunt", "vulnerability_assessment"},
		{"web", "comprehensive", "vulnerability_assessment"},
		{"cloud", "", "aws_security_assessment"},
		{"", "ctf-web", "ctf_web_challenge"},
		{"", "ad", "active_directory_assessment"},
	}
	for _, tc := range cases {
		if got := SelectPatternKey(tc.tt, tc.obj); got != tc.want {
			t.Fatalf("%s/%s: got %q want %q", tc.tt, tc.obj, got, tc.want)
		}
	}
}

func TestOptimizeParametersWithProfile_toolBranches(t *testing.T) {
	d := DefaultDecisionEngine()
	pWeb := BuildTargetProfile("https://x", "web", []string{"php"}, "", nil, 0)
	pAPI := TargetProfile{Target: "https://api", TargetType: "api"}

	sql := d.OptimizeParametersWithProfile(pWeb, "sqlmap", nil, OptimizeContext{Aggressive: true})
	if sql["additional_args"] == "" {
		t.Fatal("sqlmap args empty")
	}
	ffuf := d.OptimizeParametersWithProfile(pAPI, "ffuf", nil, OptimizeContext{Stealth: true})
	if ffuf["match_codes"] == "" || ffuf["additional_args"] == "" {
		t.Fatalf("ffuf api stealth: %v", ffuf)
	}
	hydra := d.OptimizeParametersWithProfile(pWeb, "hydra", nil, OptimizeContext{})
	if hydra["service"] != "ssh" {
		t.Fatalf("hydra service: %q", hydra["service"])
	}

	nmapAdv := d.OptimizeParametersWithProfile(pWeb, "nmap-advanced", nil, OptimizeContext{Stealth: true})
	if nmapAdv["timing"] != "T2" {
		t.Fatalf("nmap-advanced stealth: %v", nmapAdv)
	}
	nmapAdvAgg := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "ip"}, "nmap-advanced", nil, OptimizeContext{})
	if nmapAdvAgg["os_detection"] != "true" {
		t.Fatalf("nmap-advanced aggressive defaults: %v", nmapAdvAgg)
	}

	enum := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "ip"}, "enum4linux-ng", nil, OptimizeContext{})
	if enum["shares"] != "true" {
		t.Fatalf("enum4linux-ng: %v", enum)
	}
	cloud := d.OptimizeParametersWithProfile(TargetProfile{}, "prowler", nil, OptimizeContext{})
	if cloud["provider"] != "aws" {
		t.Fatalf("prowler: %v", cloud)
	}
	legacy := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "web"}, "httpx", nil, OptimizeContext{})
	if legacy["additional_args"] != "-silent" {
		t.Fatalf("httpx legacy: %q", legacy["additional_args"])
	}
}
