package normalize

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

// Smoke: pipeline import path still forwards to pkg SOT.
func TestForwardNormalizeIOC(t *testing.T) {
	got, ok := NormalizeIOC(domain.IOC{Type: domain.IOCIP, Value: "203.0.113.1"})
	if !ok || got.Value != "203.0.113.1" {
		t.Fatalf("got %#v ok=%v", got, ok)
	}
}
