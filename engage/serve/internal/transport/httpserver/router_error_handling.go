package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/recovery"
)

func registerErrorHandling(mux *http.ServeMux) {
	h := recovery.Default()
	mux.HandleFunc("GET /api/error-handling/statistics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"success":     true,
			"recoverable": []string{"timeout", "not_found", "rate_limit", "network_unreachable", "target_unreachable", "invalid_parameters"},
			"max_retries": map[string]int{
				"timeout": 3, "rate_limit": 3, "network_unreachable": 3,
				"not_found": 1, "default": 2,
			},
			"note": "in-process recovery; statistics are static schema (no persistent history)",
		})
	})
	mux.HandleFunc("GET /api/error-handling/fallback-chains", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"success":      true,
			"alternatives": h.Alternatives(),
		})
	})
	mux.HandleFunc("GET /api/error-handling/alternative-tools", func(w http.ResponseWriter, r *http.Request) {
		tool := r.URL.Query().Get("tool")
		errType := recovery.ErrorType(r.URL.Query().Get("error_type"))
		if errType == "" {
			errType = recovery.TypeTimeout
		}
		alt := h.SuggestAlternative(tool, errType)
		writeJSON(w, http.StatusOK, map[string]any{
			"success":    true,
			"tool":       tool,
			"error_type": errType,
			"alternative": alt,
		})
	})
	postJSON(mux, "POST /api/error-handling/classify-error", func(r *http.Request, body map[string]any) (any, int) {
		msg := toString(body["error"])
		if msg == "" {
			msg = toString(body["message"])
		}
		t := h.Classify(msg)
		return map[string]any{
			"success":     true,
			"error_type":  t,
			"recoverable": h.Recoverable(t),
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/error-handling/parameter-adjustments", func(r *http.Request, body map[string]any) (any, int) {
		tool := toString(body["tool"])
		errType := recovery.ErrorType(toString(body["error_type"]))
		if errType == "" {
			errType = h.Classify(toString(body["error"]))
		}
		params := map[string]string{}
		if raw, ok := body["params"].(map[string]any); ok {
			for k, v := range raw {
				params[k] = toString(v)
			}
		}
		return map[string]any{
			"success":    true,
			"tool":       tool,
			"error_type": errType,
			"params":     h.AdjustParams(tool, errType, params),
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/error-handling/test-recovery", func(r *http.Request, body map[string]any) (any, int) {
		msg := toString(body["error"])
		t := h.Classify(msg)
		return map[string]any{
			"success":     true,
			"error_type":  t,
			"recoverable": h.Recoverable(t),
			"max_retries": h.MaxRetries(t),
			"backoff_sec": int(h.BackoffDelay(1).Seconds()),
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/error-handling/execute-with-recovery", func(r *http.Request, body map[string]any) (any, int) {
		msg := toString(body["error"])
		tool := toString(body["tool"])
		t := h.Classify(msg)
		return map[string]any{
			"success":     true,
			"tool":        tool,
			"error_type":  t,
			"recoverable": h.Recoverable(t),
			"alternative": h.SuggestAlternative(tool, t),
			"note":        "use POST /api/tools/{name} — runner applies recovery automatically",
		}, http.StatusOK
	})
}
