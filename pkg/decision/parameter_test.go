package decision

import "testing"

func TestOptimizeParametersWithProfile_rustscan_modes(t *testing.T) {
	d := DefaultDecisionEngine()
	profile := TargetProfile{Target: "10.0.0.1", TargetType: "ip"}

	stealth := d.OptimizeParametersWithProfile(profile, "rustscan", nil, OptimizeContext{Stealth: true})
	if stealth["batch_size"] != "500" || stealth["ulimit"] != "1000" {
		t.Fatalf("stealth rustscan: %v", stealth)
	}

	aggressive := d.OptimizeParametersWithProfile(profile, "rustscan", nil, OptimizeContext{Aggressive: true})
	if aggressive["batch_size"] != "8000" || aggressive["ulimit"] != "10000" {
		t.Fatalf("aggressive rustscan: %v", aggressive)
	}

	comprehensive := d.OptimizeParametersWithProfile(profile, "rustscan", nil, OptimizeContext{Objective: "comprehensive"})
	if comprehensive["scripts"] != "true" {
		t.Fatalf("comprehensive scripts: %v", comprehensive)
	}
}

func TestOptimizeParametersWithProfile_masscan_rates(t *testing.T) {
	d := DefaultDecisionEngine()
	profile := TargetProfile{TargetType: "ip"}

	stealth := d.OptimizeParametersWithProfile(profile, "masscan", nil, OptimizeContext{Stealth: true})
	if stealth["rate"] != "100" {
		t.Fatalf("stealth rate: %q", stealth["rate"])
	}
	aggressive := d.OptimizeParametersWithProfile(profile, "masscan", nil, OptimizeContext{Aggressive: true})
	if aggressive["rate"] != "10000" {
		t.Fatalf("aggressive rate: %q", aggressive["rate"])
	}
	defaultRate := d.OptimizeParametersWithProfile(profile, "masscan", nil, OptimizeContext{})
	if defaultRate["rate"] != "1000" || defaultRate["ports"] != "1-65535" {
		t.Fatalf("default masscan: %v", defaultRate)
	}
}

func TestOptimizeParametersWithProfile_nuclei_quick(t *testing.T) {
	d := DefaultDecisionEngine()
	p := BuildTargetProfile("https://wp.example", "web", []string{"wordpress"}, "wordpress", nil, 0)
	out := d.OptimizeParametersWithProfile(p, "nuclei", map[string]string{}, OptimizeContext{Quick: true})
	if out["severity"] != "critical,high" {
		t.Fatalf("quick severity: %q", out["severity"])
	}
	if out["tags"] == "" {
		t.Fatal("expected wordpress tag from profile")
	}
}

func TestOptimizeParametersWithProfile_gobuster_phpExtensions(t *testing.T) {
	d := DefaultDecisionEngine()
	p := BuildTargetProfile("https://php.example", "web", []string{"php", "apache"}, "", nil, 0)
	out := d.OptimizeParametersWithProfile(p, "gobuster", map[string]string{}, OptimizeContext{})
	if out["mode"] != "dir" {
		t.Fatalf("mode: %q", out["mode"])
	}
	if out["additional_args"] == "" || out["additional_args"] == "-x html,php,txt,js -t 20" {
		t.Fatalf("expected php-specific extensions, got %q", out["additional_args"])
	}
}

func TestOptimizeParametersWithProfile_setsTargetFromProfile(t *testing.T) {
	d := DefaultDecisionEngine()
	p := TargetProfile{Target: "scan.me", TargetType: "web"}
	out := d.OptimizeParametersWithProfile(p, "httpx", map[string]string{}, OptimizeContext{})
	if out["target"] != "scan.me" {
		t.Fatalf("target: %q", out["target"])
	}
}
