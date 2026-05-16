package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	domainreport "github.com/butbeautifulv/veil/engage/serve/internal/domain/report"
	"github.com/butbeautifulv/veil/pkg/auth"
)

func parseFindings(raw any) []domainreport.Finding {
	if raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var findings []domainreport.Finding
	if err := json.Unmarshal(b, &findings); err != nil {
		return nil
	}
	return findings
}

func postJSON(mux *http.ServeMux, pattern string, fn func(*http.Request, map[string]any) (any, int)) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		if body == nil {
			body = map[string]any{}
		}
		res, code := fn(r, body)
		writeJSON(w, code, res)
	})
}

func subject(r *http.Request) string {
	if sub, ok := auth.SubjectFromContext(r.Context()); ok {
		return sub.Sub
	}
	return ""
}

func toInt(v any, def int) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return def
	}
}

func toBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "1", "true", "yes", "on":
			return true
		}
	case float64:
		return t != 0
	case int:
		return t != 0
	}
	return false
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
