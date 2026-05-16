package httpserver

import (
	"net/http"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
)

func registerProcessRoutes(mux *http.ServeMux, c *components.APIComponents) {
	mux.HandleFunc("GET /api/process/resource-usage", func(w http.ResponseWriter, r *http.Request) {
		dash := c.Processes.Dashboard()
		out := map[string]any{
			"timestamp":       dash["timestamp"],
			"uptime_sec":      int(time.Since(c.StartedAt).Seconds()),
			"total_processes": dash["total_processes"],
			"running":         dash["running"],
			"system_load":     dash["system_load"],
		}
		writeJSON(w, http.StatusOK, out)
	})
}
