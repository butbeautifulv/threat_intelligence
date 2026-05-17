package config

import (
	"os"
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	os.Unsetenv("VEIL_GATEWAY_LISTEN")
	os.Unsetenv("VEIL_GRAPH_API_URL")
	os.Unsetenv("VEIL_ENGAGE_API_URL")
	cfg := Load()
	if cfg.ListenAddr != ":8080" {
		t.Fatalf("listen %q", cfg.ListenAddr)
	}
	if cfg.GraphAPIURL != "http://127.0.0.1:8090" {
		t.Fatalf("graph %q", cfg.GraphAPIURL)
	}
	if cfg.EngageAPIURL != "http://127.0.0.1:8890" {
		t.Fatalf("engage %q", cfg.EngageAPIURL)
	}
}
