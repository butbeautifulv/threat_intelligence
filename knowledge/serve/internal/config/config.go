package config

import (
	"os"
	"strings"

	"github.com/butbeautifulv/veil/pkg/auth"
)

type Config struct {
	ListenAddr string
	Env        string
	Neo4j      Neo4jConfig
	Auth       auth.Config
	MCPHTTP    MCPHTTPConfig
	Security   SecurityConfig
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

func LoadAPI() *Config {
	return loadBase(getenv("API_LISTEN", ":8090"), getenv("API_ENV", "local"))
}

// LoadMCP loads config for the MCP stdio server.
func LoadMCP() *Config {
	return loadBase("", getenv("MCP_ENV", "local"))
}

func loadBase(listen, env string) *Config {
	return &Config{
		ListenAddr: listen,
		Env:        env,
		Neo4j: Neo4jConfig{
			URI:      getenv("NEO4J_URI", "neo4j://localhost:7687"),
			Username: getenv("NEO4J_USER", "neo4j"),
			Password: getenv("NEO4J_PASS", "neo4jpassword"),
			Database: getenv("NEO4J_DB", "neo4j"),
		},
		Auth:     auth.LoadConfigFromEnv(),
		MCPHTTP:  loadMCPHTTPFromEnv(),
		Security: loadSecurityFromEnv(env),
	}
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
