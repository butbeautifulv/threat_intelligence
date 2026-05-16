package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
)

func registerCommand(mux *http.ServeMux, c *components.APIComponents) {
	postJSON(mux, "POST /api/command", func(r *http.Request, body map[string]any) (any, int) {
		cmd, _ := body["command"].(string)
		useCache := true
		if body["use_cache"] == false {
			useCache = false
		}
		return c.Command.Run(r.Context(), cmd, useCache, c.Cache), http.StatusOK
	})
}
