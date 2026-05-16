package audit

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExportWebhook_hmac(t *testing.T) {
	var gotSig string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Engage-Signature")
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	events := []Event{{Tool: "nmap_scan", Target: "127.0.0.1", Success: true}}
	if err := ExportWebhook(context.Background(), srv.URL, "secret", events); err != nil {
		t.Fatal(err)
	}
	if gotSig == "" || len(gotSig) < 10 {
		t.Fatalf("expected signature, got %q", gotSig)
	}
}
