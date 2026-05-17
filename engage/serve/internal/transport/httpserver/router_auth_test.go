package httpserver

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	authmw "github.com/butbeautifulv/veil/engage/serve/internal/auth/middleware"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/static"
)

func TestGetToolsCatalog_requiresBearerWhenAuthEnabled(t *testing.T) {
	cfg := config.LoadAPI()
	cfg.CatalogPath = filepath.Join("..", "..", "..", "catalog", "tools.live.yaml")
	cfg.FilesDir = t.TempDir()
	cfg.RunnerWork = t.TempDir()
	c, err := components.InitAPI(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	v := static.New("pentest-catalog-token", "runner-1", []string{"veil-engage-runner"})
	stack := auth.NewStack(v, auth.Config{
		Enabled:          true,
		RBACEnabled:      true,
		RoleEngageRunner: "veil-engage-runner",
	})
	c.Auth = stack

	mux := http.NewServeMux()
	Register(mux, c)
	handler := authmw.Auth(stack, false, cfg.Security, mux)

	t.Run("no_token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("valid_token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
		req.Header.Set("Authorization", "Bearer pentest-catalog-token")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
	})
}
