package config

import (
	"os"
	"strings"
)

type Config struct {
	ListenAddr string
	Env        string
	Neo4j      Neo4jConfig
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

func Load() *Config {
	return &Config{
		ListenAddr: envOr("API_LISTEN_ADDR", ":8090"),
		Env:        envOr("API_ENV", "local"),
		Neo4j: Neo4jConfig{
			URI:      envOr("NEO4J_URI", "neo4j://localhost:7687"),
			Username: envOr("NEO4J_USER", "neo4j"),
			Password: envOr("NEO4J_PASS", "neo4jpassword"),
			Database: envOr("NEO4J_DB", "neo4j"),
		},
	}
}

func envOr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
