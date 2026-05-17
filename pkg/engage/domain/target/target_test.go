package target

import "testing"

func TestTargetJSONFields(t *testing.T) {
	tg := Target{Value: "example.com", Kind: "host"}
	if tg.Value == "" {
		t.Fatal("value must be set")
	}
}
