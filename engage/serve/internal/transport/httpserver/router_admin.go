package httpserver

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	domainjob "github.com/butbeautifulv/veil/pkg/engage/domain/job"
	"github.com/butbeautifulv/veil/engage/serve/internal/telemetry"
)

func registerAdmin(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /api/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		if c.Cache != nil {
			writeJSON(w, http.StatusOK, c.Cache.Stats())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": 0})
	})
	mux.HandleFunc("POST /api/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		cleared := 0
		if c.Cache != nil {
			cleared = c.Cache.Clear()
		}
		writeJSON(w, http.StatusOK, map[string]any{"cleared": cleared})
	})
	mux.HandleFunc("GET /api/telemetry", func(w http.ResponseWriter, r *http.Request) {
		out := map[string]any{
			"uptime_sec":      int(time.Since(c.StartedAt).Seconds()),
			"tools_enabled":   len(c.Tools.List()),
			"processes_total": len(c.Processes.List()),
		}
		running := 0
		for _, p := range c.Processes.List() {
			if p.Status == "running" {
				running++
			}
		}
		out["processes_running"] = running
		if c.Cache != nil {
			stats := c.Cache.Stats()
			out["cache_entries"] = stats["entries"]
			if entries, ok := stats["entries"].(int); ok {
				telemetry.SetCacheEntries(entries)
			}
		}
		if c.Jobs != nil {
			if n, err := c.Jobs.CountByStatus(domainjob.StatusPending); err == nil {
				out["jobs_pending"] = n
				telemetry.SetJobsPending(n)
			}
			if n, err := c.Jobs.CountByStatus(domainjob.StatusRunning); err == nil {
				out["jobs_running"] = n
			}
		}
		writeJSON(w, http.StatusOK, out)
	})
	mux.HandleFunc("GET /api/processes/list", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"processes": c.Processes.List()})
	})
	mux.HandleFunc("GET /api/processes/status/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		rec, ok := c.Processes.Get(pid)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})
	mux.HandleFunc("POST /api/processes/terminate/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		if err := c.Processes.Terminate(r.Context(), pid); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"terminated": pid})
	})
	mux.HandleFunc("POST /api/processes/pause/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		if err := c.Processes.Pause(pid); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"paused": pid})
	})
	mux.HandleFunc("POST /api/processes/resume/{pid}", func(w http.ResponseWriter, r *http.Request) {
		pid, _ := strconv.Atoi(r.PathValue("pid"))
		if err := c.Processes.Resume(pid); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"resumed": pid})
	})
	mux.HandleFunc("GET /api/processes/dashboard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, c.Processes.Dashboard())
	})
	mux.HandleFunc("GET /api/audit/recent", func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if c.AuditReader == nil {
			writeJSON(w, http.StatusOK, map[string]any{"events": []any{}})
			return
		}
		events, err := c.AuditReader.Recent(limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	})
	mux.HandleFunc("GET /api/audit/export", func(w http.ResponseWriter, r *http.Request) {
		if c.AuditReader == nil {
			http.Error(w, "audit store not configured", http.StatusServiceUnavailable)
			return
		}
		var since time.Time
		if raw := strings.TrimSpace(r.URL.Query().Get("since")); raw != "" {
			if t, err := time.Parse(time.RFC3339, raw); err == nil {
				since = t
			}
		}
		data, err := c.AuditReader.ExportNDJSON(since)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	postJSON(mux, "POST /api/audit/export-webhook", func(r *http.Request, body map[string]any) (any, int) {
		url := toString(body["url"])
		if url == "" {
			url = c.AuditWebhookURL
		}
		if url == "" {
			return map[string]any{"error": "webhook url required"}, http.StatusBadRequest
		}
		if c.AuditReader == nil {
			return map[string]any{"error": "audit store not configured"}, http.StatusServiceUnavailable
		}
		events, err := c.AuditReader.Recent(500)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusInternalServerError
		}
		secret := c.AuditWebhookSecret
		if s := toString(body["secret"]); s != "" {
			secret = s
		}
		if err := audit.ExportWebhook(r.Context(), url, secret, events); err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadGateway
		}
		return map[string]any{"exported": len(events), "url": url}, http.StatusOK
	})
	if c.MetricsEnabled {
		mux.Handle("GET /metrics", telemetry.Handler())
	}
}
