package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
)

// Register attaches unified veil-api routes: composite health and upstream proxies.
func Register(mux *http.ServeMux, cfg config.Config, client *http.Client) {
	mux.HandleFunc("GET /health", compositeHealth(cfg, client))
	mux.Handle("/v1/", newPrefixProxy(cfg.GraphAPIURL, "graph"))
	mux.Handle("/api/", newPrefixProxy(cfg.EngageAPIURL, "engage"))
}
