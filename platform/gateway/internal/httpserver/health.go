package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
	"github.com/butbeautifulv/veil/pkg/api"
)

// BackendStatus is one upstream /health probe result.
type BackendStatus struct {
	OK       bool           `json:"ok"`
	URL      string           `json:"url"`
	Status   int              `json:"status,omitempty"`
	Service  string           `json:"service,omitempty"`
	Detail   map[string]any   `json:"detail,omitempty"`
	Error    string           `json:"error,omitempty"`
}

func probeBackend(ctx context.Context, client *http.Client, baseURL string) BackendStatus {
	baseURL = strings.TrimRight(baseURL, "/")
	st := BackendStatus{URL: baseURL}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	resp, err := client.Do(req)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	defer resp.Body.Close()
	st.Status = resp.StatusCode
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(body) > 0 {
			st.Error = strings.TrimSpace(string(body))
		} else {
			st.Error = resp.Status
		}
		return st
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		st.Error = "invalid health json"
		return st
	}
	st.Detail = parsed
	if v, ok := parsed["ok"].(bool); ok {
		st.OK = v
	} else {
		st.OK = true
	}
	if svc, ok := parsed["service"].(string); ok {
		st.Service = svc
	}
	return st
}

// compositeHealth mirrors pkg/api.RegisterHealth (ok, service, WriteJSON) and adds graph/engage probes.
func compositeHealth(cfg config.Config, client *http.Client) http.HandlerFunc {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		graph := probeBackend(ctx, client, cfg.GraphAPIURL)
		engage := probeBackend(ctx, client, cfg.EngageAPIURL)
		ok := graph.OK && engage.OK
		status := http.StatusOK
		if !ok {
			status = http.StatusServiceUnavailable
		}
		body := map[string]any{
			"ok":      ok,
			"service": "veil-api",
			"graph":   graph,
			"engage":  engage,
		}
		api.WriteJSON(w, status, body)
	}
}
