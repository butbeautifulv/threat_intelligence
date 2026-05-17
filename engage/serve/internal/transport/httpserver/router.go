package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tooldispatch"
	"github.com/butbeautifulv/veil/pkg/api"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// Register attaches Veil Engage HTTP routes to mux.
func Register(mux *http.ServeMux, c *components.APIComponents) {
	api.RegisterHealth(mux, "veil-engage", map[string]any{"tool_count": c.Registry.Count()})

	mux.HandleFunc("GET /api/tools", func(w http.ResponseWriter, r *http.Request) {
		list := c.Tools.List()
		out := make([]map[string]any, 0, len(list))
		for _, s := range list {
			out = append(out, map[string]any{
				"name": s.Name, "category": string(s.Category), "description": s.Description,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"tools": out})
	})

	mux.HandleFunc("POST /api/tools/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var req contract.ToolRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}
		if rejectBlockedTarget(w, c, req.Target, name) {
			return
		}
		sub := subject(r)
		body, err := c.ToolDispatch.DispatchRequest(r.Context(), sub, name, req)
		if err != nil {
			status := http.StatusInternalServerError
			var de *tooldispatch.DispatchError
			if errors.As(err, &de) && de.NotFound {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]any{"success": false, "tool": name, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, body)
	})

	registerJobs(mux, c)
	registerIntel(mux, c)
	registerWorkflows(mux, c)
	registerCTF(mux, c)
	registerVulnIntel(mux, c)
	registerErrorHandling(mux)
	registerProcessRoutes(mux, c)
	registerBrowser(mux, c)
	registerVisual(mux, c)
	registerFiles(mux, c)
	registerCommand(mux, c)
	registerPayloads(mux, c)
	registerAdmin(mux, c)
	registerPlaybooks(mux, c)
}
