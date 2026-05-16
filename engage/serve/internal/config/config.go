package config

import (
	"os"
	"strings"

	"github.com/butbeautifulv/veil/pkg/auth"
)

type Config struct {
	ListenAddr  string
	Env         string
	Auth        auth.Config
	Security    SecurityConfig
	CatalogPath string
	RunnerWork  string
	VeilAPI     VeilAPIConfig
	MCPHTTP     MCPHTTPConfig
}

type VeilAPIConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	TokenURL     string
}

type MCPHTTPConfig struct {
	Enabled   bool
	Listen    string
	Path      string
	BindLocal bool
}

func LoadAPI() *Config {
	return loadBase(getenv("ENGAGE_API_LISTEN", ":8890"), getenv("ENGAGE_ENV", "local"))
}

func LoadMCP() *Config {
	return loadBase("", getenv("ENGAGE_ENV", "local"))
}

func loadBase(listen, env string) *Config {
	return &Config{
		ListenAddr:  listen,
		Env:         env,
		Auth:        loadAuthFromEnv(),
		Security:    LoadSecurityForEnv(env),
		CatalogPath: getenv("ENGAGE_CATALOG_PATH", "catalog/tools.yaml"),
		RunnerWork:  getenv("ENGAGE_RUNNER_WORKDIR", "/tmp/engage"),
		VeilAPI: VeilAPIConfig{
			BaseURL:      getenv("ENGAGE_VEIL_API_URL", "http://localhost:8090"),
			ClientID:     getenv("ENGAGE_VEIL_CLIENT_ID", ""),
			ClientSecret: getenv("ENGAGE_VEIL_CLIENT_SECRET", ""),
			TokenURL:     getenv("ENGAGE_VEIL_TOKEN_URL", ""),
		},
		MCPHTTP: MCPHTTPConfig{
			Enabled:   envBool("ENGAGE_MCP_HTTP_ENABLED", false),
			Listen:    getenv("ENGAGE_MCP_HTTP_LISTEN", ":8892"),
			Path:      getenv("ENGAGE_MCP_HTTP_PATH", "/mcp"),
			BindLocal: envBool("ENGAGE_MCP_HTTP_BIND_LOCAL", false),
		},
	}
}

func loadAuthFromEnv() auth.Config {
	c := auth.LoadConfigFromEnv()
	// Engage defaults when using engage-specific env prefix
	if v := getenv("ENGAGE_AUTH_ENABLED", ""); v != "" {
		c.Enabled = envBool("ENGAGE_AUTH_ENABLED", c.Enabled)
	}
	return c
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
