package middleware

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/pkg/api"
	"github.com/butbeautifulv/veil/pkg/auth"
)

func Auth(stack *auth.Stack, strict bool, sec config.SecurityConfig, next http.Handler) http.Handler {
	return api.AuthMiddleware(stack, strict, sec.Prod, auth.PermEngageToolRun, next)
}
