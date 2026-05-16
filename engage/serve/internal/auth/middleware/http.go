package middleware

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/httpmiddleware"
)

func Auth(stack *auth.Stack, strict bool, sec config.SecurityConfig, next http.Handler) http.Handler {
	return httpmiddleware.Auth(stack, strict, sec.Prod, auth.PermEngageToolRun, next)
}
