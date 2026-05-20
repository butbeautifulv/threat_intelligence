package decision

import "testing"

func TestOptimizeParametersWithProfile_allToolCases(t *testing.T) {
	d := DefaultDecisionEngine()
	p := TargetProfile{Target: "10.0.0.1", TargetType: "web"}
	ctx := OptimizeContext{}

	tools := []string{
		"enum4linux-ng", "enum4linux", "autorecon", "ghidra", "pwntools", "ropper", "angr",
		"prowler", "scout-suite", "kube-hunter", "trivy", "checkov",
		"httpx", "feroxbuster", "nikto", "wpscan", "unknown-tool",
	}
	for _, tool := range tools {
		out := d.OptimizeParametersWithProfile(p, tool, map[string]string{}, ctx)
		if out["target"] != p.Target {
			t.Fatalf("%s target: %v", tool, out)
		}
	}
}

func TestOptimizeParametersWithProfile_nmap_defaultBranch(t *testing.T) {
	d := DefaultDecisionEngine()
	out := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "unknown"}, "nmap", nil, OptimizeContext{})
	if out["scan_type"] != "-sV" {
		t.Fatalf("default nmap: %v", out)
	}
}

func TestOptimizeParametersWithProfile_legacyNmapNuclei(t *testing.T) {
	d := DefaultDecisionEngine()
	out := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "web"}, "nuclei", map[string]string{}, OptimizeContext{})
	if out["templates"] == "" {
		t.Fatal("expected web nuclei templates")
	}
}

func TestRiskAndConfidence_helpers(t *testing.T) {
	if riskFromAttackSurface(9) != "critical" || riskFromAttackSurface(1) != "minimal" {
		t.Fatal("risk levels")
	}
	p := TargetProfile{TargetType: "custom", Technologies: []string{"unknown"}}
	if calculateConfidence(p) < 0.6 {
		t.Fatal("confidence baseline")
	}
	if calculateAttackSurface(TargetProfile{TargetType: "custom"}) != 3.0 {
		t.Fatal("unknown type base")
	}
}

func TestOptimizeParametersWithProfile_nmap_sqlmap_ffuf_branches(t *testing.T) {
	d := DefaultDecisionEngine()
	stealth := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "ip"}, "nmap", map[string]string{"additional_args": "-T4"}, OptimizeContext{Stealth: true})
	if !containsSubstr(stealth["additional_args"], "-T2") {
		t.Fatalf("nmap stealth: %v", stealth)
	}
	php := BuildTargetProfile("https://x", "web", []string{"php"}, "", nil, 0)
	sql := d.OptimizeParametersWithProfile(php, "sqlmap", nil, OptimizeContext{})
	if sql["additional_args"] == "" {
		t.Fatal("sqlmap php")
	}
	dotnet := BuildTargetProfile("https://x", "web", []string{"dotnet"}, "", nil, 0)
	sqlNet := d.OptimizeParametersWithProfile(dotnet, "sqlmap", nil, OptimizeContext{})
	if sqlNet["additional_args"] == "" {
		t.Fatal("sqlmap dotnet")
	}
	sqlDef := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "web"}, "sqlmap", map[string]string{}, OptimizeContext{})
	if sqlDef["additional_args"] == "" {
		t.Fatal("sqlmap default batch")
	}
	ffufAPI := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "api"}, "ffuf", nil, OptimizeContext{})
	if ffufAPI["match_codes"] == "" {
		t.Fatal("ffuf api")
	}
	ffufWeb := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "web"}, "ffuf", nil, OptimizeContext{})
	if ffufWeb["additional_args"] == "" {
		t.Fatal("ffuf web")
	}
	joomla := BuildTargetProfile("https://x", "web", []string{"joomla"}, "", nil, 0)
	nuc := d.OptimizeParametersWithProfile(joomla, "nuclei", nil, OptimizeContext{})
	if nuc["tags"] == "" {
		t.Fatal("nuclei joomla tag")
	}
	gob := d.OptimizeParametersWithProfile(BuildTargetProfile("https://x", "web", []string{"java"}, "", nil, 0), "gobuster", nil, OptimizeContext{})
	if gob["additional_args"] == "" {
		t.Fatal("gobuster java")
	}
	dotGob := d.OptimizeParametersWithProfile(BuildTargetProfile("https://x", "web", []string{"asp.net"}, "", nil, 0), "gobuster", nil, OptimizeContext{})
	if dotGob["additional_args"] == "" {
		t.Fatal("gobuster dotnet")
	}
	drupal := d.OptimizeParametersWithProfile(BuildTargetProfile("https://x", "web", []string{"drupal"}, "", nil, 0), "nuclei", nil, OptimizeContext{})
	if drupal["tags"] == "" {
		t.Fatal("drupal tag")
	}
}

func TestSelectPatternKey_allObjectives(t *testing.T) {
	cases := map[string]string{
		"wireless": "wireless_assessment", "mobile": "mobile_app_assessment",
		"phishing": "phishing_assessment", "supply-chain": "supply_chain_assessment",
		"bugbounty-recon": "bug_bounty_reconnaissance", "osint": "bug_bounty_reconnaissance",
		"crypto": "ctf_crypto_challenge", "forensics": "ctf_forensics_challenge",
		"pwn": "ctf_pwn_challenge", "file-upload": "bug_bounty_vulnerability_hunting",
		"high-impact": "bug_bounty_high_impact", "business-logic": "bug_bounty_vulnerability_hunting",
		"vulnerability": "vulnerability_assessment", "binary": "binary_exploitation",
	}
	for tt, obj := range map[string]string{
		"web": "unknown-obj", "cloud": "multi-cloud", "binary": "",
	} {
		_ = SelectPatternKey(tt, obj)
	}
	for _, obj := range []string{
		"recon", "reconnaissance", "vuln", "vulnerability", "vuln_hunt", "vuln-hunt",
		"bugbounty-recon", "osint", "bugbounty-high", "high-impact", "business-logic",
		"file-upload", "ctf", "ctf-web", "pwn", "ctf-pwn", "forensics", "crypto",
		"binary", "exploit", "ad", "active-directory", "supply-chain", "wireless",
		"mobile", "phishing", "comprehensive", "multi-cloud",
	} {
		for _, tt := range []string{"web", "api", "ip", "cloud", "binary", "other"} {
			_ = SelectPatternKey(tt, obj)
		}
	}
	for obj, want := range cases {
		if got := SelectPatternKey("web", obj); got != want {
			t.Fatalf("%s: got %q want %q", obj, got, want)
		}
	}
}

func TestEffectivenessTable(t *testing.T) {
	d := DefaultDecisionEngine()
	if len(d.effectivenessTable("web")) == 0 {
		t.Fatal("web table empty")
	}
	if len(d.effectivenessTable("not-a-type")) == 0 {
		t.Fatal("unknown table empty")
	}
}

func TestBoostValue(t *testing.T) {
	if boostValue(nil, "nmap") != 0 {
		t.Fatal("nil boost")
	}
	if boostValue(map[string]float64{"nmap": 2}, "nmap") != 2 {
		t.Fatal("boost value")
	}
}

func TestCompareRankedDesc(t *testing.T) {
	if compareRankedDesc(rankedTool{"a", 1}, rankedTool{"b", 2}) <= 0 {
		t.Fatal("desc order")
	}
	if compareRankedDesc(rankedTool{"a", 2}, rankedTool{"b", 2}) != 0 {
		t.Fatal("equal scores")
	}
}

func TestRankTools_delegatesToBoost(t *testing.T) {
	d := DefaultDecisionEngine()
	if got := d.RankTools("web", []string{"nuclei", "nmap"}); len(got) != 2 || got[0] != "nuclei" {
		t.Fatalf("RankTools: %v", got)
	}
}

func TestRankToolsWithBoost_and_Score_unknown(t *testing.T) {
	d := DefaultDecisionEngine()
	ranked := d.RankToolsWithBoost("web", []string{"nmap", "unknown-tool"}, map[string]float64{"unknown-tool": 2.0})
	if len(ranked) != 2 {
		t.Fatalf("ranked: %v", ranked)
	}
	ranked2 := d.RankToolsWithBoost("web", []string{"nikto", "nmap", "nuclei"}, map[string]float64{"nikto": 5.0})
	ranked3 := d.RankToolsWithBoost("unknown", []string{"tool-a", "tool-b", "tool-c"}, nil)
	if len(ranked3) != 3 {
		t.Fatalf("ranked3: %v", ranked3)
	}
	ranked4 := d.RankToolsWithBoost("web", []string{"nmap", "nuclei"}, map[string]float64{})
	if len(ranked4) != 2 {
		t.Fatalf("ranked4: %v", ranked4)
	}
	_ = d.RankToolsWithBoost("web", []string{"nuclei", "nmap", "nikto"}, map[string]float64{"nmap": 0, "nikto": 0})
	if len(d.RankToolsWithBoost("web", nil, nil)) != 0 {
		t.Fatal("nil candidates")
	}
	if len(d.RankToolsWithBoost("web", []string{}, nil)) != 0 {
		t.Fatal("empty candidates")
	}
	if d.RankToolsWithBoost("web", []string{"nuclei"}, nil)[0] != "nuclei" {
		t.Fatal("single candidate")
	}
	if len(ranked2) != 3 {
		t.Fatalf("ranked2: %v", ranked2)
	}
	if d.Score("unknown-type", "nonexistent-tool-xyz") != 0.5 {
		t.Fatalf("unknown tool score: %v", d.Score("unknown-type", "nonexistent-tool-xyz"))
	}
}

func TestCalculateConfidence_allBonuses(t *testing.T) {
	p := TargetProfile{
		TargetType: "web", CMS: "wp",
		Technologies: []string{"nginx"},
		IPAddresses:  []string{"10.0.0.1"},
	}
	if calculateConfidence(p) != 1.0 {
		t.Fatalf("got %v", calculateConfidence(p))
	}
}

func TestFilterStealthTools_truncates(t *testing.T) {
	ids := []string{"amass", "subfinder", "httpx", "nuclei", "extra1", "extra2"}
	out := FilterStealthTools(ids)
	if len(out) != 4 {
		t.Fatalf("got %v", out)
	}
	only := FilterStealthTools([]string{"nmap", "masscan"})
	if len(only) != 0 {
		t.Fatalf("got %v", only)
	}
}

func TestApplyLegacyParameterDefaults_tools(t *testing.T) {
	d := DefaultDecisionEngine()
	for _, tc := range []struct {
		tool, want string
	}{
		{"feroxbuster", "-q"},
		{"nikto", "123bde"},
		{"wpscan", "enumerate"},
	} {
		out := d.OptimizeParametersWithProfile(TargetProfile{TargetType: "web"}, tc.tool, map[string]string{}, OptimizeContext{})
		if out["additional_args"] == "" || !containsSubstr(out["additional_args"], tc.want) {
			t.Fatalf("%s: %v", tc.tool, out)
		}
	}
	out := map[string]string{}
	applyLegacyParameterDefaults("web", "nmap", out)
	if out["scan_type"] != "-sV" {
		t.Fatalf("legacy nmap: %v", out)
	}
	out2 := map[string]string{}
	applyLegacyParameterDefaults("web", "nuclei", out2)
	if out2["templates"] == "" {
		t.Fatalf("legacy nuclei: %v", out2)
	}
}

func containsSubstr(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && indexSub(s, sub) >= 0))
}
func indexSub(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestOptimizeParameters_includesTarget(t *testing.T) {
	d := DefaultDecisionEngine()
	out := d.OptimizeParameters("ip", "nmap", map[string]string{"target": "host"})
	if out["target"] != "host" {
		t.Fatalf("got %v", out)
	}
}
