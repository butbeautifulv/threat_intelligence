package bugbounty

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReconWorkflow_golden(t *testing.T) {
	m := NewManager()
	wf := m.CreateReconnaissance(Target{Domain: "example.com"})
	if len(wf.Phases) < 4 {
		t.Fatalf("phases %d", len(wf.Phases))
	}
	names := make([]string, len(wf.Phases))
	for i, p := range wf.Phases {
		names[i] = p.Name
	}
	b, err := os.ReadFile(filepath.Join("testdata", "recon_workflow.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		PhasesMin     int      `json:"phases_min"`
		ToolsCountMin int      `json:"tools_count_min"`
		PhaseNames    []string `json:"phase_names"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	if len(wf.Phases) < spec.PhasesMin {
		t.Fatalf("phases %d < %d", len(wf.Phases), spec.PhasesMin)
	}
	if wf.ToolsCount < spec.ToolsCountMin {
		t.Fatalf("tools %d", wf.ToolsCount)
	}
	for i, want := range spec.PhaseNames {
		if names[i] != want {
			t.Fatalf("phase[%d] %q != %q", i, names[i], want)
		}
	}
}
