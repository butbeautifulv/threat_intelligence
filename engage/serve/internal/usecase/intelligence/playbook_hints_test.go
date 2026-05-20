package intelligence

import (
	"context"
	"encoding/json"
	"testing"
)

type mockPlaybookVeil struct {
	mockGraphReader
	byTechnique map[string]json.RawMessage
}

func (m *mockPlaybookVeil) PlaybooksByTechnique(_ context.Context, tid string) (json.RawMessage, error) {
	return m.byTechnique[tid], nil
}

func (m *mockPlaybookVeil) PlaybookRecommendTools(_ context.Context, _, tid string) (json.RawMessage, error) {
	return m.byTechnique[tid], nil
}

func TestAttachPlaybookHints_extractsTechniqueIDs(t *testing.T) {
	veil := &mockPlaybookVeil{
		byTechnique: map[string]json.RawMessage{
			"T1059.001": json.RawMessage(`{"technique_id":"T1059.001","index_count":1}`),
			"T1003":     json.RawMessage(`{"technique_id":"T1003","index_count":2}`),
		},
	}
	out := map[string]any{}
	attachPlaybookHints(context.Background(), veil, out,
		"investigate T1059.001 on host", "objective T1003")
	hints, ok := out["playbook_hints"].(map[string]json.RawMessage)
	if !ok || len(hints) != 2 {
		t.Fatalf("hints %#v", out["playbook_hints"])
	}
	if _, ok := hints["T1059.001"]; !ok {
		t.Fatal("missing T1059.001")
	}
	if _, ok := hints["T1003"]; !ok {
		t.Fatal("missing T1003")
	}
}

func TestCollectAttackTechniqueIDs_capsAtFive(t *testing.T) {
	blob := "T1001 T1002 T1003 T1004 T1005 T1006 T1007"
	ids := collectAttackTechniqueIDs(blob)
	if len(ids) != 5 {
		t.Fatalf("got %d ids", len(ids))
	}
}
