package intelligence

import "testing"

func TestDecisionEngine_OptimizeParameters_nmap(t *testing.T) {
	d := DefaultDecisionEngine()
	out := d.OptimizeParameters("ip", "nmap", map[string]string{})
	if out["scan_type"] != "-sV" {
		t.Fatalf("scan_type: %q", out["scan_type"])
	}
}

func TestDecisionEngine_RankTools(t *testing.T) {
	d := DefaultDecisionEngine()
	ranked := d.RankTools("web", []string{"nikto", "nuclei", "nmap"})
	if ranked[0] != "nuclei" {
		t.Fatalf("expected nuclei first, got %v", ranked)
	}
}
