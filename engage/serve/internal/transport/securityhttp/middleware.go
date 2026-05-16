package securityhttp

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/config"
)

type responseWriter struct {
	http.ResponseWriter
	stripped bool
}

func (w *responseWriter) WriteHeader(code int) {
	if !w.stripped {
		w.ResponseWriter.Header().Del("Server")
		w.stripped = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func Harden(sec config.SecurityConfig, maxBody int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		if sec.Prod {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			if len(sec.CORSAllowedOrigins) == 0 {
				http.Error(w, "cors not allowed", http.StatusForbidden)
				return
			}
			allowed := false
			for _, o := range sec.CORSAllowedOrigins {
				if o == origin && o != "" {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "cors not allowed", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		if maxBody > 0 && r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		}
		next.ServeHTTP(&responseWriter{ResponseWriter: w}, r)
	})
}

func HTTPServerTimeouts() (readHeader, read, write, idle int) {
	return 10, 30, 120, 120
}
