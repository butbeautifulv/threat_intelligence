package ctf

import (
	"testing"

	domaintool "github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
)

func TestCreateChallengeWorkflow_categories(t *testing.T) {
	mgr := NewManager()
	cases := []struct {
		category string
		minTools int
	}{
		{"web", 3},
		{"pwn", 2},
		{"crypto", 1},
		{"forensics", 2},
		{"misc", 0},
	}
	for _, tc := range cases {
		ch := Challenge{Name: "test", Category: tc.category, Description: "sql injection web app", Difficulty: "medium"}
		_ = ch.Validate(false)
		suggested := mgr.Tools.SuggestTools(ch.Description, ch.Category)
		wf := mgr.CreateChallengeWorkflow(ch, suggested)
		if wf.Category != ch.Category {
			t.Fatalf("%s category %q", tc.category, wf.Category)
		}
		if len(wf.WorkflowSteps) == 0 {
			t.Fatalf("%s: no steps", tc.category)
		}
		if len(suggested) < tc.minTools && tc.minTools > 0 {
			t.Fatalf("%s: tools %v", tc.category, suggested)
		}
	}
}

func TestSuggestTools_sqlKeyword(t *testing.T) {
	m := NewToolManager()
	tools := m.SuggestTools("test sql injection in login", "web")
	found := false
	for _, id := range tools {
		if id == "sqlmap" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sqlmap in %v", tools)
	}
}

func TestResolveTools_registry(t *testing.T) {
	reg := tools.NewRegistry([]domaintool.Spec{{Name: "nmap_scan"}, {Name: "httpx_probe"}})
	m := NewToolManager()
	got := m.ResolveTools([]string{"nmap", "httpx"}, reg)
	if len(got) != 2 {
		t.Fatalf("resolved %v", got)
	}
}
