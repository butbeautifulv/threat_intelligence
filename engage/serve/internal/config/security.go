package config

import (
	"errors"
	"os"
	"strings"
)

type SecurityConfig struct {
	RequireAuth        bool
	MCPHTTPAuthStrict  bool
	Prod               bool
	AllowRawCommand    bool
	CORSAllowedOrigins []string
	APIBodyLimit       int64
	MCPBodyLimit       int64
}

func LoadSecurityForEnv(env string) SecurityConfig {
	prod := strings.EqualFold(strings.TrimSpace(env), "prod")
	cors := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	var origins []string
	if cors != "" {
		for _, o := range strings.Split(cors, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins = append(origins, o)
			}
		}
	}
	allowRaw := envBool("ENGAGE_ALLOW_RAW_COMMAND", false)
	if prod || envBool("ENGAGE_DENY_RAW_COMMAND", false) {
		allowRaw = false
	}
	return SecurityConfig{
		RequireAuth:        envBool("VEIL_REQUIRE_AUTH", false),
		MCPHTTPAuthStrict:  envBool("ENGAGE_MCP_HTTP_AUTH_STRICT", false),
		Prod:               prod,
		AllowRawCommand:    allowRaw,
		CORSAllowedOrigins: origins,
		APIBodyLimit:       4 << 20,
		MCPBodyLimit:       8 << 20,
	}
}

func ValidateSecurity(sec SecurityConfig, authEnabled bool) error {
	if sec.RequireAuth && !authEnabled {
		return errors.New("VEIL_REQUIRE_AUTH=1 requires AUTH_ENABLED=1")
	}
	return nil
}
