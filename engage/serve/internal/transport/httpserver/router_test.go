package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	domainjob "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
)

func initTestAPI(t *testing.T) *components.APIComponents {
	t.Helper()
	cfg := config.LoadAPI()
	cfg.CatalogPath = filepath.Join("..", "..", "..", "catalog", "tools.live.yaml")
	cfg.FilesDir = t.TempDir()
	cfg.RunnerWork = t.TempDir()
	c, err := components.InitAPI(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestHealth(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestPostJob_withParameters(t *testing.T) {
	if _, err := exec.LookPath("nmap"); err != nil {
		t.Skip("nmap not on PATH")
	}
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)

	body, _ := json.Marshal(map[string]any{
		"tool":   "nmap_scan",
		"target": "127.0.0.1",
		"parameters": map[string]string{
			"scan_type": "-sn",
			"ports":     "",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("missing job id")
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		req2 := httptest.NewRequest(http.MethodGet, "/api/jobs/"+id, nil)
		rr2 := httptest.NewRecorder()
		mux.ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusOK {
			t.Fatalf("get job status %d", rr2.Code)
		}
		var job map[string]any
		_ = json.Unmarshal(rr2.Body.Bytes(), &job)
		if st, _ := job["status"].(string); st == "done" || st == "failed" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("job did not finish")
}

func TestProcesses_afterToolRun(t *testing.T) {
	if _, err := exec.LookPath("nmap"); err != nil {
		t.Skip("nmap not on PATH")
	}
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)

	body, _ := json.Marshal(map[string]any{
		"target": "127.0.0.1",
		"parameters": map[string]string{
			"scan_type": "-sn",
			"ports":     "",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/tools/nmap_scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	req2 := httptest.NewRequest(http.MethodGet, "/api/processes/list", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("processes status %d", rr2.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	procs, _ := resp["processes"].([]any)
	if len(procs) == 0 {
		t.Fatalf("expected process records after tool run, tool status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSmartScan_maxTools(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{
		"target":    "https://example.com",
		"max_tools": 1,
		"async":     true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/intelligence/smart-scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	selected, _ := resp["tools_selected"].([]any)
	if len(selected) > 1 {
		t.Fatalf("expected max 1 tool, got %v", selected)
	}
}

func TestFiles_createAndList(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{"filename": "test.txt", "content": "data"})
	req := httptest.NewRequest(http.MethodPost, "/api/files/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("create status %d", rr.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/api/files/list?directory=.", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("list status %d", rr2.Code)
	}
}

func TestCache_stats(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	req := httptest.NewRequest(http.MethodGet, "/api/cache/stats", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestJobs_listAndCancel(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{"tool": "nmap_scan", "target": "127.0.0.1"})
	req := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("create job %d", rr.Code)
	}
	var created map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	id, _ := created["id"].(string)

	reqList := httptest.NewRequest(http.MethodGet, "/api/jobs?status=pending", nil)
	rrList := httptest.NewRecorder()
	mux.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("list %d", rrList.Code)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := c.Jobs.Get(id)
		if !ok {
			break
		}
		if got.Status == domainjob.StatusPending {
			reqCancel := httptest.NewRequest(http.MethodPost, "/api/jobs/"+id+"/cancel", nil)
			rrCancel := httptest.NewRecorder()
			mux.ServeHTTP(rrCancel, reqCancel)
			if rrCancel.Code != http.StatusOK {
				t.Fatalf("cancel %d body %s", rrCancel.Code, rrCancel.Body.String())
			}
			return
		}
		if got.Status == domainjob.StatusDone || got.Status == domainjob.StatusFailed || got.Status == domainjob.StatusCancelled {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestPayloads_generate(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{"type": "buffer", "size": 8, "pattern": "X", "filename": "p.txt"})
	req := httptest.NewRequest(http.MethodPost, "/api/payloads/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
}

func TestTelemetry_expanded(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	req := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if _, ok := resp["uptime_sec"]; !ok {
		t.Fatalf("telemetry: %v", resp)
	}
}

func TestAssessmentReport_shape(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{
		"target":    "https://example.com",
		"max_tools": 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/intelligence/assessment-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["summary_report"]; !ok {
		t.Fatalf("missing summary_report: %v", resp)
	}
	if _, ok := resp["findings"]; !ok {
		t.Fatalf("missing findings: %v", resp)
	}
}

func TestSummaryReport_withFindings(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{
		"target": "https://example.com",
		"findings": []map[string]any{
			{"title": "x", "severity": "medium", "description": "d", "target": "https://example.com"},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/visual/summary-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	findings, _ := resp["findings"].([]any)
	if len(findings) != 1 {
		t.Fatalf("findings: %v", resp["findings"])
	}
}

func TestAuditRecent_emptyWhenNoStore(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	req := httptest.NewRequest(http.MethodGet, "/api/audit/recent?limit=10", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestAuditRecent_withStore(t *testing.T) {
	c := initTestAPI(t)
	dir := t.TempDir()
	store, err := audit.NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	c.AuditStore = store
	c.AuditReader = store
	c.Audit = audit.NewWithStore(nil, store)
	_ = store.Append(audit.Event{Subject: "u", Tool: "nmap_scan", Target: "127.0.0.1", Success: true})
	mux := http.NewServeMux()
	Register(mux, c)
	req := httptest.NewRequest(http.MethodGet, "/api/audit/recent?limit=10", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	events, _ := resp["events"].([]any)
	if len(events) != 1 {
		t.Fatalf("events: %v", resp)
	}
}

func TestExportReport_pdf(t *testing.T) {
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{
		"target": "https://example.com",
		"findings": []map[string]any{
			{"title": "x", "severity": "high", "description": "d", "target": "https://example.com"},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/visual/export-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["pdf_base64"] == nil {
		t.Fatalf("missing pdf_base64: %v", resp)
	}
}

func TestCommand_allowlistedBinary(t *testing.T) {
	if _, err := exec.LookPath("nmap"); err != nil {
		t.Skip("nmap not on PATH")
	}
	c := initTestAPI(t)
	mux := http.NewServeMux()
	Register(mux, c)
	body, _ := json.Marshal(map[string]any{"command": "nmap --version"})
	req := httptest.NewRequest(http.MethodPost, "/api/command", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}
