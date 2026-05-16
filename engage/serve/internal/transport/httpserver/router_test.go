package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
)

func TestHealth(t *testing.T) {
	cfg := config.LoadAPI()
	cfg.CatalogPath = "../../catalog/tools.live.yaml"
	c, err := components.InitAPI(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	Register(mux, c)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}
