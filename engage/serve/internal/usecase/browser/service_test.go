package browser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInspect_parsesSidecar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inspect" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"page_info": map[string]any{
				"title": "Example",
				"forms": []any{map[string]any{"action": "/login"}},
			},
			"security_analysis": map[string]any{
				"security_score": 80,
				"total_issues":   1,
			},
			"technologies": []string{"nginx"},
		})
	}))
	defer srv.Close()

	svc := &Service{BaseURL: srv.URL, Client: srv.Client()}
	out := svc.Inspect(context.Background(), InspectRequest{URL: "https://example.com"})
	if !out.Success {
		t.Fatalf("success false: %s", out.Error)
	}
	if out.PageInfo == nil {
		t.Fatal("missing page_info")
	}
	if out.SecurityAnalysis == nil {
		t.Fatal("missing security_analysis")
	}
	if score, ok := out.SecurityAnalysis["security_score"].(float64); !ok || score <= 0 {
		t.Fatalf("security_score unexpected: %#v", out.SecurityAnalysis["security_score"])
	}
	if len(out.Forms) != 1 {
		t.Fatalf("forms normalized: %#v", out.Forms)
	}
	if action, _ := out.Forms[0]["action"].(string); action != "/login" {
		t.Fatalf("form action got %q", action)
	}
}
