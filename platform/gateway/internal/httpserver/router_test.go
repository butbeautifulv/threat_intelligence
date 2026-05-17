package httpserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
)

func TestCompositeHealth_ok(t *testing.T) {
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"service":"veil-api"}`)
	}))
	defer graph.Close()
	engage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"service":"veil-engage","tool_count":1}`)
	}))
	defer engage.Close()

	mux := http.NewServeMux()
	Register(mux, config.Config{GraphAPIURL: graph.URL, EngageAPIURL: engage.URL}, graph.Client())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != true || body["service"] != "veil-api" {
		t.Fatalf("body %v", body)
	}
	graphSt := body["graph"].(map[string]any)
	if graphSt["ok"] != true {
		t.Fatalf("graph %v", graphSt)
	}
}

func TestCompositeHealth_degraded(t *testing.T) {
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer graph.Close()
	engage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"service":"veil-engage"}`)
	}))
	defer engage.Close()

	mux := http.NewServeMux()
	Register(mux, config.Config{GraphAPIURL: graph.URL, EngageAPIURL: engage.URL}, graph.Client())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != false {
		t.Fatalf("body %v", body)
	}
}

func TestProxyV1(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/categories" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"categories":[]}`)
	}))
	defer up.Close()

	mux := http.NewServeMux()
	Register(mux, config.Config{GraphAPIURL: up.URL, EngageAPIURL: up.URL}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestProxyAPI(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tools" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"tools":[]}`)
	}))
	defer up.Close()

	mux := http.NewServeMux()
	Register(mux, config.Config{GraphAPIURL: up.URL, EngageAPIURL: up.URL}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}
