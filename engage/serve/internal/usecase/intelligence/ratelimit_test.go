package intelligence

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseRateLimitOutput(t *testing.T) {
	if ok, _ := parseRateLimitOutput("HTTP/1.1 429 Too Many Requests"); !ok {
		t.Fatal("expected 429 detection")
	}
	if ok, _ := parseRateLimitOutput("ok"); ok {
		t.Fatal("unexpected detection")
	}
}

func TestProbeHTTPRateLimit_429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	probe := probeHTTPRateLimit(context.Background(), srv.URL)
	if !probe.Detected || probe.Source != "http" {
		t.Fatalf("probe %+v", probe)
	}
}

func TestProbeHTTPRateLimit_retryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	probe := probeHTTPRateLimit(context.Background(), srv.URL)
	if !probe.Detected {
		t.Fatalf("probe %+v", probe)
	}
}

func TestProbeRateLimit_emptyTarget(t *testing.T) {
	s := &Service{}
	probe := s.ProbeRateLimit(context.Background(), "subj", "")
	if probe.Detected || probe.Source != "none" {
		t.Fatalf("probe %+v", probe)
	}
}
