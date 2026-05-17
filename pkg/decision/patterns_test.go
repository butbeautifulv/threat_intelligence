package decision

import "testing"

func TestAttackPatterns_count(t *testing.T) {
	p := AttackPatterns()
	if len(p) < 20 {
		t.Fatalf("expected >= 20 patterns, got %d", len(p))
	}
}

func TestAttackPatterns_eachHasSteps(t *testing.T) {
	for key, steps := range AttackPatterns() {
		if len(steps) == 0 {
			t.Fatalf("pattern %q has no steps", key)
		}
	}
}

func TestSelectPatternKey_binaryAndCloud(t *testing.T) {
	if got := SelectPatternKey("binary", "exploit"); got != "binary_exploitation" {
		t.Fatalf("binary: %q", got)
	}
	if got := SelectPatternKey("cloud", "multi-cloud"); got != "multi_cloud_assessment" {
		t.Fatalf("multi-cloud: %q", got)
	}
}

func TestAttackPatterns_webRecon_hasParams(t *testing.T) {
	steps := AttackPatterns()["web_reconnaissance"]
	for _, st := range steps {
		if st.Tool == "gobuster" && st.Params["mode"] != "dir" {
			t.Fatalf("gobuster params: %v", st.Params)
		}
	}
}
