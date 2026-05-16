package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	domainjob "github.com/butbeautifulv/veil/engage/serve/internal/domain/job"
)

func registerJobs(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("POST /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Tool       string            `json:"tool"`
			Target     string            `json:"target"`
			Parameters map[string]string `json:"parameters"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
			return
		}
		j, err := c.Jobs.Enqueue(body.Tool, body.Target, subject(r), body.Parameters)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, j)
	})
	mux.HandleFunc("GET /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		list, err := c.Jobs.List(domainjob.Status(status), limit)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"jobs": list})
	})
	mux.HandleFunc("GET /api/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		j, ok := c.Jobs.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusOK, j)
	})
	mux.HandleFunc("POST /api/jobs/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := c.Jobs.Cancel(id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		j, _ := c.Jobs.Get(id)
		writeJSON(w, http.StatusOK, j)
	})
}
