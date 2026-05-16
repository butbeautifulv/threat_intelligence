package middleware

import (
	"net/http"

	"github.com/butbeautifulv/veil/graph/serve/internal/config"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/httpmiddleware"
)

// Auth wraps handlers with JWT Bearer auth and optional RBAC.
func Auth(stack *auth.Stack, strict bool, sec config.SecurityConfig, next http.Handler) http.Handler {
	return httpmiddleware.Auth(stack, strict, sec.Prod, auth.PermGraphRead, next)
}
