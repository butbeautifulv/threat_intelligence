package veilgraph

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetJSON_withoutOAuth_noAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	raw, err := c.GetJSON(context.Background(), "/v1/categories")
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("body %s", raw)
	}
	if gotAuth != "" {
		t.Fatalf("expected no Authorization, got %q", gotAuth)
	}
}
