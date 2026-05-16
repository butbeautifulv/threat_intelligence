package cve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNVDClient_FetchCVE(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "nvd_cve_sample.json"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cveId") == "" {
			t.Fatal("expected cveId query")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	client := NewNVDClient(srv.URL, srv.Client())
	entry, err := client.FetchCVE(context.Background(), "CVE-2021-44228")
	if err != nil {
		t.Fatal(err)
	}
	if entry.CVEID != "CVE-2021-44228" {
		t.Fatalf("cve id %q", entry.CVEID)
	}
	if entry.Severity != "CRITICAL" {
		t.Fatalf("severity %q", entry.Severity)
	}
	if entry.CVSSScore < 9.9 {
		t.Fatalf("cvss %v", entry.CVSSScore)
	}
	if !contains(entry.Description, "Log4j") {
		t.Fatal("expected log4j description")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
