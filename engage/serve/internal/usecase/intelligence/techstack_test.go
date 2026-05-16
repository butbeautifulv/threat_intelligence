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

func TestMatchHeaderSignatures_nginx(t *testing.T) {
	h := http.Header{}
	h.Set("Server", "nginx/1.24.0")
	tech := MatchHeaderSignatures(h)
	if len(tech) == 0 || tech[0] != TechNginx {
		t.Fatalf("expected nginx, got %v", tech)
	}
}

func TestMatchContentSignatures_wordpress(t *testing.T) {
	body := `<html><link href="/wp-content/themes/foo">`
	tech := MatchContentSignatures(body)
	found := false
	for _, t := range tech {
		if t == TechWordPress {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected wordpress in %v", tech)
	}
}
