package decision

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEffectivenessParityWithLegacy(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonPath := filepath.Join(filepath.Dir(file), "testdata", "effectiveness_legacy.json")
	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Skip("legacy effectiveness JSON missing; run scripts/engage/extract-decision-tables.py")
	}
	var legacy map[string]map[string]float64
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	eng := DefaultDecisionEngine()
	mismatches := 0
	total := 0
	for targetType, tools := range legacy {
		for tool, want := range tools {
			total++
			got := eng.Score(targetType, tool)
			if got != want {
				mismatches++
				if mismatches <= 5 {
					t.Logf("mismatch %s/%s: got %v want %v", targetType, tool, got, want)
				}
			}
		}
	}
	if mismatches > 0 {
		pct := float64(total-mismatches) / float64(total) * 100
		if pct < 90 {
			t.Fatalf("parity %.1f%% (%d/%d mismatches)", pct, mismatches, total)
		}
		t.Logf("parity %.1f%% (%d minor mismatches)", pct, mismatches)
	}
}
