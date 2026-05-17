package config

import "os"

// Config holds unified veil-api gateway settings.
type Config struct {
	ListenAddr   string
	GraphAPIURL  string
	EngageAPIURL string
}

// Load reads gateway configuration from the environment.
func Load() Config {
	return Config{
		ListenAddr:   getenv("VEIL_GATEWAY_LISTEN", ":8080"),
		GraphAPIURL:  getenv("VEIL_GRAPH_API_URL", "http://127.0.0.1:8090"),
		EngageAPIURL: getenv("VEIL_ENGAGE_API_URL", "http://127.0.0.1:8890"),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
