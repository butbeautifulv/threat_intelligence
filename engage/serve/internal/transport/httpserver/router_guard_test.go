package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/security"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

func TestPostTool_metadataProbeForbiddenBeforeRun(t *testing.T) {
	c := initTestAPI(t)
	c.Tools.TargetGuard = security.TargetGuardOff
	mux := http.NewServeMux()
	Register(mux, c)

	body, _ := json.Marshal(map[string]any{
		"target": "http://169.254.169.254/latest/meta-data/",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/tools/amass_passive_scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var resp contract.ToolRunResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Success {
		t.Fatal("expected success=false")
	}
	if !strings.Contains(resp.Error, "ENGAGE_TARGET_GUARD") {
		t.Fatalf("error %q", resp.Error)
	}
	if strings.Contains(strings.ToLower(resp.Output), "ami-id") {
		t.Fatal("response must not leak metadata probe output")
	}
}

func TestPostJob_metadataProbeForbidden(t *testing.T) {
	c := initTestAPI(t)
	c.Tools.TargetGuard = security.TargetGuardOff
	mux := http.NewServeMux()
	Register(mux, c)

	body, _ := json.Marshal(map[string]any{
		"tool":   "nmap_scan",
		"target": "http://169.254.169.254/latest/meta-data/",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}
