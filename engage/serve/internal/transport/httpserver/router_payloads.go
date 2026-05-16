package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/payloads"
)

func registerPayloads(mux *http.ServeMux, c *components.APIComponents) {
	postJSON(mux, "POST /api/payloads/generate", func(r *http.Request, body map[string]any) (any, int) {
		req := payloads.Request{
			Type:     toString(body["type"]),
			Size:     toInt(body["size"], 1024),
			Pattern:  toString(body["pattern"]),
			Filename: toString(body["filename"]),
		}
		if req.Pattern == "" {
			req.Pattern = "A"
		}
		res, err := payloads.Generate(c.Files, req)
		if err != nil {
			return map[string]any{"error": err.Error()}, http.StatusBadRequest
		}
		return res, http.StatusOK
	})
}
