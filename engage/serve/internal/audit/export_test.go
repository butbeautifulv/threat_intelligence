package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExportWebhook(t *testing.T) {
	var got int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got++
		if r.Header.Get("X-Engage-Signature") == "" {
			t.Fatal("expected signature header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	events := []Event{{Tool: "nmap_scan", Target: "127.0.0.1", At: time.Now().UTC(), Success: true}}
	if err := ExportWebhook(context.Background(), srv.URL, "secret", events); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("expected 1 post, got %d", got)
	}
}
