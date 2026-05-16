package intelligence

import (
	"context"
	"net/http"
	"testing"
)

func TestAllTechnologies_count(t *testing.T) {
	if len(AllTechnologies()) != 15 {
		t.Fatalf("expected 15 technologies, got %d", len(AllTechnologies()))
	}
}

func TestDetectTechnologies_wordpress(t *testing.T) {
	h := http.Header{}
	h.Set("Server", "Apache")
	h.Set("X-Powered-By", "PHP/8.1")
	tech := DetectTechnologies(context.Background(), "https://example.com/wp-admin", h, "")
	found := map[string]bool{}
	for _, t := range tech {
		found[string(t)] = true
	}
	if !found["wordpress"] && !found["php"] && !found["apache"] {
		t.Fatalf("expected wordpress/php/apache, got %v", tech)
	}
}

func TestTechStackBoost_wordpress(t *testing.T) {
	b := techStackBoost([]Technology{TechWordPress})
	if b["wpscan"] < 0.2 {
		t.Fatalf("wpscan boost: %v", b)
	}
}
