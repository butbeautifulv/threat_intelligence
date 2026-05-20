package report

import "testing"

func TestDefaultBranding(t *testing.T) {
	b := DefaultBranding()
	if b.Organization != "Veil Engage" {
		t.Fatalf("Organization: %q", b.Organization)
	}
	if b.Classification != "CONFIDENTIAL" {
		t.Fatalf("Classification: %q", b.Classification)
	}
	if b.Footer == "" {
		t.Fatal("expected default footer")
	}
}
