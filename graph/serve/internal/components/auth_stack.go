package components

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/veil/graph/serve/internal/auth"
	"github.com/butbeautifulv/veil/graph/serve/internal/auth/keycloak"
)

func newAuthStack(ctx context.Context, cfg auth.Config) (*auth.Stack, error) {
	if !cfg.Enabled {
		return auth.NewStack(nil, cfg), nil
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("AUTH_ENABLED=1 requires KEYCLOAK_ISSUER")
	}
	v, err := keycloak.NewVerifier(ctx, cfg.Issuer, cfg.Audience, cfg.ClientID)
	if err != nil {
		return nil, err
	}
	return auth.NewStack(v, cfg), nil
}
