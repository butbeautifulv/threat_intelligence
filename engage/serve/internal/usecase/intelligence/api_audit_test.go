package intelligence

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestComprehensiveAPIAudit_minimal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	s := &Service{}
	out := s.ComprehensiveAPIAudit(context.Background(), "", ComprehensiveAPIAuditRequest{
		BaseURL: srv.URL,
	})
	if out["success"] != true {
		t.Fatalf("expected success: %v", out)
	}
	performed, _ := out["tests_performed"].([]string)
	if len(performed) < 1 {
		t.Fatalf("expected tests_performed: %v", performed)
	}
}

func TestComprehensiveAPIAudit_requiresBaseURL(t *testing.T) {
	s := &Service{}
	out := s.ComprehensiveAPIAudit(context.Background(), "", ComprehensiveAPIAuditRequest{})
	if out["success"] != false {
		t.Fatal("expected failure without base_url")
	}
}

func TestPhaseJWTAnalysis(t *testing.T) {
	// alg=none style header {"alg":"none"} -> eyJhbGciOiJub25lIn0
	tok := "eyJhbGciOiJub25lIn0.eyJzdWIiOiIxIn0."
	out := JWTAnalysis(tok)
	if out["findings_count"].(int) < 1 {
		t.Fatalf("expected jwt findings: %v", out)
	}
}
