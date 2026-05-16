package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBrowserProxy_Exec(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/exec" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stdout": "ok", "stderr": "", "exit_code": 0,
		})
	}))
	defer srv.Close()

	t.Setenv("ENGAGE_BROWSER_URL", srv.URL)
	proxy := NewBrowserProxyFromEnv()
	res := proxy.Exec(context.Background(), "https://example.com", []string{"https://example.com"})
	if res.ExitCode != 0 || res.Stdout != "ok" {
		t.Fatalf("res: %+v", res)
	}
}

func TestIsBrowserBinary(t *testing.T) {
	if !IsBrowserBinary("browser") {
		t.Fatal("expected browser")
	}
}

func TestExecutor_browserProxy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"stdout": "navigated\n", "exit_code": 0})
	}))
	defer srv.Close()
	t.Setenv("ENGAGE_BROWSER_URL", srv.URL)
	ex := &Executor{}
	res := ex.Run(context.Background(), "browser", []string{"https://example.com"}, 0, nil)
	if res.ExitCode != 0 {
		t.Fatalf("%+v", res)
	}
	os.Unsetenv("ENGAGE_BROWSER_URL")
}
