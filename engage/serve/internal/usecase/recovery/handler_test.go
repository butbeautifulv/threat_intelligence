package recovery

import "testing"

func TestClassify_timeout(t *testing.T) {
	h := Default()
	if got := h.Classify("operation timed out"); got != TypeTimeout {
		t.Fatalf("got %s", got)
	}
}

func TestSuggestAlternative_nuclei(t *testing.T) {
	h := Default()
	alt := h.SuggestAlternative("nuclei_scan", TypeNotFound)
	if alt != "jaeles" && alt != "nikto" {
		t.Fatalf("alt: %q", alt)
	}
}

func TestRecoverable(t *testing.T) {
	h := Default()
	if !h.Recoverable(TypeTimeout) {
		t.Fatal("timeout should be recoverable")
	}
	if h.Recoverable(TypePermission) {
		t.Fatal("permission should not be recoverable")
	}
}
