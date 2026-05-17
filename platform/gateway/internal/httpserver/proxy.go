package httpserver

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/butbeautifulv/veil/pkg/api"
)

func newPrefixProxy(rawBase string, label string) http.Handler {
	u, err := url.Parse(rawBase)
	if err != nil {
		log.Fatalf("gateway %s proxy: invalid base URL %q: %v", label, rawBase, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		orig(req)
		if req.Header.Get("X-Forwarded-Host") == "" {
			req.Header.Set("X-Forwarded-Host", req.Host)
		}
		if req.Header.Get("X-Forwarded-Proto") == "" {
			if req.TLS != nil {
				req.Header.Set("X-Forwarded-Proto", "https")
			} else {
				req.Header.Set("X-Forwarded-Proto", "http")
			}
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		api.WriteError(w, http.StatusBadGateway, errors.New(label+" upstream unavailable: "+err.Error()))
	}
	return proxy
}
