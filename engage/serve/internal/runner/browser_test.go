package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestBrowserProxy_Inspect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inspect" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"page_info": map[string]any{
				"forms": []any{},
			},
			"security_analysis": map[string]any{"security_score": 90},
		})
	}))
	defer srv.Close()

	t.Setenv("DISCOVERY_BROWSER_URL", srv.URL)
	proxy := NewBrowserProxyFromEnv()
	res := proxy.Inspect(context.Background(), BrowserInspectOpts{Target: "https://example.com"})
	if res.ExitCode != 0 {
		t.Fatalf("res: %+v", res)
	}
	if !strings.Contains(res.Stdout, "security_score") {
		t.Fatalf("stdout: %s", res.Stdout)
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
	t.Setenv("DISCOVERY_BROWSER_URL", srv.URL)
	ex := &Executor{}
	res := ex.Run(context.Background(), "browser", []string{"https://example.com"}, 0, nil)
	if res.ExitCode != 0 {
		t.Fatalf("%+v", res)
	}
	os.Unsetenv("DISCOVERY_BROWSER_URL")
}
