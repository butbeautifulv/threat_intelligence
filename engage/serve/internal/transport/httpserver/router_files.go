package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
)

func registerFiles(mux *http.ServeMux, c *components.APIComponents) {
	if c.Files == nil {
		return
	}
	postJSON(mux, "POST /api/files/create", func(r *http.Request, body map[string]any) (any, int) {
		res, err := c.Files.Create(toString(body["filename"]), toString(body["content"]), body["binary"] == true)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
	postJSON(mux, "POST /api/files/modify", func(r *http.Request, body map[string]any) (any, int) {
		res, err := c.Files.Modify(toString(body["filename"]), toString(body["content"]), body["append"] == true)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
	postJSON(mux, "POST /api/files/delete", func(r *http.Request, body map[string]any) (any, int) {
		res, err := c.Files.Delete(toString(body["filename"]))
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
	mux.HandleFunc("GET /api/files/list", func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("directory")
		res, err := c.Files.List(dir)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
}
