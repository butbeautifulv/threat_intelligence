package mcpserver

import (
	"encoding/json"
	"testing"
)

func TestNegotiateProtocol(t *testing.T) {
	p, _ := json.Marshal(map[string]any{"protocolVersion": protocol20250326})
	if got := negotiateProtocol(p); got != protocol20250326 {
		t.Fatalf("got %q want %q", got, protocol20250326)
	}
	if got := negotiateProtocol(nil); got != defaultProtocol {
		t.Fatalf("default: got %q", got)
	}
}
