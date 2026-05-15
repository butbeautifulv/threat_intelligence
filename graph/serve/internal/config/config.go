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

func LoadAPI() *Config {
	return &Config{
		ListenAddr: getenv("API_LISTEN", ":8090"),
		Env:        getenv("API_ENV", "local"),
		Neo4j: Neo4jConfig{
			URI:      getenv("NEO4J_URI", "neo4j://localhost:7687"),
			Username: getenv("NEO4J_USER", "neo4j"),
			Password: getenv("NEO4J_PASS", "neo4jpassword"),
			Database: getenv("NEO4J_DB", "neo4j"),
		},
	}
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
