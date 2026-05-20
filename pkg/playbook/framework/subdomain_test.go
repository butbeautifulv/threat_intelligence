package framework

import "testing"

func TestLoadSubdomains(t *testing.T) {
	subs, err := LoadSubdomains()
	if err != nil {
		t.Skip(err)
	}
	if len(subs) < 20 {
		t.Fatalf("subdomains: %d", len(subs))
	}
}

func TestSkillsForTechnique_T1059(t *testing.T) {
	ids, err := SkillsForTechnique("T1059.001")
	if err != nil {
		t.Skip(err)
	}
	if len(ids) == 0 {
		t.Fatal("expected skills for T1059.001")
	}
}
