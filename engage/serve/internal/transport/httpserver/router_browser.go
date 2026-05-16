package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
)

func registerBrowser(mux *http.ServeMux, c *components.APIComponents) {
	if c.Browser == nil {
		return
	}
	postJSON(mux, "POST /api/browser/inspect", func(r *http.Request, body map[string]any) (any, int) {
		url := toString(body["url"])
		if url == "" {
			url = toString(body["target"])
		}
		if url == "" {
			return map[string]any{"success": false, "error": "url or target required"}, http.StatusBadRequest
		}
		params := map[string]string{}
		for k, v := range body {
			if s, ok := v.(string); ok {
				params[k] = s
			}
		}
		out := c.Browser.Inspect(r.Context(), browserInspectReqFromBody(url, body, params))
		if !out.Success {
			return out, http.StatusOK
		}
		return out, http.StatusOK
	})
}

func browserInspectReqFromBody(url string, body map[string]any, params map[string]string) browser.InspectRequest {
	req := browser.InspectFromParams(url, params)
	if body != nil {
		if v, ok := body["wait_time"]; ok {
			switch n := v.(type) {
			case float64:
				req.WaitTime = int(n)
			case int:
				req.WaitTime = n
			}
		}
		if body["active_tests"] == true {
			req.ActiveTests = true
		}
	}
	return req
}
