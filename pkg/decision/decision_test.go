package decision

import "testing"

func TestDecisionEngine_OptimizeParameters_nmap(t *testing.T) {
	d := DefaultDecisionEngine()
	out := d.OptimizeParameters("ip", "nmap", map[string]string{})
	if out["scan_type"] != "-sS -O" {
		t.Fatalf("ip scan_type: %q", out["scan_type"])
	}
	outWeb := d.OptimizeParameters("web", "nmap", map[string]string{})
	if outWeb["scan_type"] != "-sV -sC" {
		t.Fatalf("web scan_type: %q", outWeb["scan_type"])
	}
}

func TestDecisionEngine_OptimizeParameters_gobuster(t *testing.T) {
	d := DefaultDecisionEngine()
	out := d.OptimizeParameters("web", "gobuster", map[string]string{})
	if out["mode"] != "dir" {
		t.Fatalf("mode: %q", out["mode"])
	}
}

func TestDecisionEngine_RankTools(t *testing.T) {
	d := DefaultDecisionEngine()
	ranked := d.RankTools("web", []string{"nikto", "nuclei", "nmap"})
	if ranked[0] != "nuclei" {
		t.Fatalf("expected nuclei first, got %v", ranked)
	}
}
