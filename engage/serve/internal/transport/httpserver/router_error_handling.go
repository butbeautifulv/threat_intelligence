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
			"recoverable": []string{"timeout", "tool_not_found", "rate_limited", "network_unreachable", "target_unreachable", "invalid_parameters"},
			"max_retries": map[string]int{
				"timeout": 3, "rate_limited": 3, "network_unreachable": 3,
				"tool_not_found": 1, "default": 2,
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
		if tool == "" {
			tool = r.URL.Query().Get("tool_name")
		}
		errType := recovery.ErrorType(r.URL.Query().Get("error_type"))
		if errType == "" {
			errType = recovery.TypeTimeout
		} else if parsed, ok := recovery.ParseErrorType(string(errType)); ok {
			errType = parsed
		}
		alt := h.SuggestAlternative(tool, errType)
		writeJSON(w, http.StatusOK, map[string]any{
			"success":     true,
			"tool":        tool,
			"tool_name":   tool,
			"error_type":  errType,
			"alternative": alt,
		})
	})
	postJSON(mux, "POST /api/error-handling/classify-error", func(r *http.Request, body map[string]any) (any, int) {
		msg := toString(body["error_message"])
		if msg == "" {
			msg = toString(body["error"])
		}
		if msg == "" {
			msg = toString(body["message"])
		}
		t := h.Classify(msg)
		strategies := h.RecoveryStrategies(t)
		out := make([]map[string]any, 0, len(strategies))
		for _, s := range strategies {
			out = append(out, map[string]any{
				"action":              s.Action,
				"parameters":          s.Parameters,
				"success_probability": s.SuccessProbability,
				"estimated_time":      s.EstimatedTime,
			})
		}
		return map[string]any{
			"success":             true,
			"error_type":          t,
			"recoverable":         h.Recoverable(t),
			"recovery_strategies": out,
			"primary_action":      h.PrimaryAction(t),
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/error-handling/parameter-adjustments", func(r *http.Request, body map[string]any) (any, int) {
		tool := toString(body["tool"])
		if tool == "" {
			tool = toString(body["tool_name"])
		}
		errTypeStr := toString(body["error_type"])
		var errType recovery.ErrorType
		if errTypeStr != "" {
			if parsed, ok := recovery.ParseErrorType(errTypeStr); ok {
				errType = parsed
			} else {
				return map[string]any{"success": false, "error": "invalid error_type: " + errTypeStr}, http.StatusBadRequest
			}
		} else {
			errType = h.Classify(toString(body["error"]))
		}
		params := map[string]string{}
		if raw, ok := body["params"].(map[string]any); ok {
			for k, v := range raw {
				params[k] = toString(v)
			}
		}
		if raw, ok := body["original_params"].(map[string]any); ok {
			for k, v := range raw {
				params[k] = toString(v)
			}
		}
		if raw, ok := body["parameters"].(map[string]any); ok {
			for k, v := range raw {
				params[k] = toString(v)
			}
		}
		adjusted := h.AdjustParams(tool, errType, params)
		return map[string]any{
			"success":         true,
			"tool":            tool,
			"tool_name":       tool,
			"error_type":      errType,
			"original_params": params,
			"adjusted_params": adjusted,
			"params":          adjusted,
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/error-handling/test-recovery", func(r *http.Request, body map[string]any) (any, int) {
		msg := toString(body["error"])
		if msg == "" {
			msg = toString(body["error_message"])
		}
		t := h.Classify(msg)
		return map[string]any{
			"success":             true,
			"error_type":          t,
			"recoverable":         h.Recoverable(t),
			"max_retries":         h.MaxRetries(t),
			"backoff_sec":         int(h.BackoffDelay(1).Seconds()),
			"primary_action":      h.PrimaryAction(t),
			"recovery_strategies": h.RecoveryStrategies(t),
		}, http.StatusOK
	})
	postJSON(mux, "POST /api/error-handling/execute-with-recovery", func(r *http.Request, body map[string]any) (any, int) {
		msg := toString(body["error"])
		if msg == "" {
			msg = toString(body["error_message"])
		}
		tool := toString(body["tool"])
		if tool == "" {
			tool = toString(body["tool_name"])
		}
		t := h.Classify(msg)
		return map[string]any{
			"success":             true,
			"tool":                tool,
			"tool_name":           tool,
			"error_type":          t,
			"recoverable":         h.Recoverable(t),
			"alternative":         h.SuggestAlternative(tool, t),
			"primary_action":      h.PrimaryAction(t),
			"recovery_strategies": h.RecoveryStrategies(t),
			"note":                "use POST /api/tools/{name} — runner applies recovery automatically",
		}, http.StatusOK
	})
}
