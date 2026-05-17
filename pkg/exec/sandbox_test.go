package exec

import (
	"os"
	"testing"
)

func TestSandboxEnabled(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		container string
		want      bool
	}{
		{"local default", "local", "", false},
		{"docker no container", "docker", "", false},
		{"docker with container", "docker", "engage-runner", true},
		{"DOCKER case", "DOCKER", "engage-runner", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sandbox{Mode: tt.mode, Container: tt.container}
			if got := s.Enabled(); got != tt.want {
				t.Fatalf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSandboxFromEnv(t *testing.T) {
	t.Setenv("ENGAGE_RUNNER_MODE", "docker")
	t.Setenv("ENGAGE_RUNNER_CONTAINER", "engage-runner")
	t.Setenv("ENGAGE_RUNNER_WORKDIR", "/tmp/engage")

	s := NewSandboxFromEnv()
	if s.Mode != "docker" {
		t.Fatalf("Mode = %q", s.Mode)
	}
	if s.Container != "engage-runner" {
		t.Fatalf("Container = %q", s.Container)
	}
	if s.WorkDir != "/tmp/engage" {
		t.Fatalf("WorkDir = %q", s.WorkDir)
	}
	if !s.Enabled() {
		t.Fatal("expected Enabled() true")
	}
}

func TestNewSandboxFromEnvEmptyMode(t *testing.T) {
	os.Unsetenv("ENGAGE_RUNNER_MODE")
	os.Unsetenv("ENGAGE_RUNNER_CONTAINER")
	t.Setenv("ENGAGE_RUNNER_WORKDIR", "")

	s := NewSandboxFromEnv()
	if s.Mode != "local" {
		t.Fatalf("Mode = %q, want local", s.Mode)
	}
	if s.Enabled() {
		t.Fatal("expected Enabled() false for default local")
	}
}
