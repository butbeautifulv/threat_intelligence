package decision

import "testing"

func TestDefaultEffectivenessTables(t *testing.T) {
	tables := defaultEffectivenessTables()
	wantTypes := []string{"web", "ip", "api", "cloud", "binary", "unknown"}
	for _, tt := range wantTypes {
		tools, ok := tables[tt]
		if !ok || len(tools) == 0 {
			t.Fatalf("missing or empty table for %q", tt)
		}
		for tool, score := range tools {
			if score <= 0 || score > 1 {
				t.Fatalf("%s/%s score %v out of range", tt, tool, score)
			}
		}
	}
	eng := DefaultDecisionEngine()
	if eng.Score("web", "nuclei") != tables["web"]["nuclei"] {
		t.Fatalf("engine should use default tables")
	}
}
