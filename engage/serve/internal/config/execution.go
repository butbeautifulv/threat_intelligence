package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	// ExecutionProfileClientNative is the default: tools run on the API/worker host process.
	ExecutionProfileClientNative = "client-native"
	// ExecutionProfileDockerExec allows ENGAGE_RUNNER_MODE=docker (docker exec into engage-runner). Use only with compose.runner / CI runners, not default prod API images.
	ExecutionProfileDockerExec = "docker-exec"
)

func loadExecutionProfile() string {
	if v := strings.TrimSpace(os.Getenv("ENGAGE_EXECUTION_PROFILE")); v != "" {
		return v
	}
	return ExecutionProfileClientNative
}

// ValidateExecutionProfile enforces runner isolation rules for ENGAGE_EXECUTION_PROFILE.
func (c *Config) ValidateExecutionProfile() error {
	p := strings.ToLower(strings.TrimSpace(c.ExecutionProfile))
	switch {
	case p == ExecutionProfileDockerExec:
		return nil
	case p == ExecutionProfileClientNative, p == "":
		mode := strings.TrimSpace(os.Getenv("ENGAGE_RUNNER_MODE"))
		if strings.EqualFold(mode, "docker") {
			return fmt.Errorf("ENGAGE_EXECUTION_PROFILE=%s forbids ENGAGE_RUNNER_MODE=docker (use %s for runner/CI overlays)",
				ExecutionProfileClientNative, ExecutionProfileDockerExec)
		}
		if strings.TrimSpace(os.Getenv("ENGAGE_RUNNER_CONTAINER")) != "" {
			return fmt.Errorf("ENGAGE_EXECUTION_PROFILE=%s forbids non-empty ENGAGE_RUNNER_CONTAINER", ExecutionProfileClientNative)
		}
		if mode != "" && !strings.EqualFold(mode, "local") {
			return fmt.Errorf("ENGAGE_EXECUTION_PROFILE=%s requires ENGAGE_RUNNER_MODE unset or local, got %q",
				ExecutionProfileClientNative, mode)
		}
		return nil
	default:
		return fmt.Errorf("unsupported ENGAGE_EXECUTION_PROFILE %q: use %q or %q",
			strings.TrimSpace(c.ExecutionProfile), ExecutionProfileClientNative, ExecutionProfileDockerExec)
	}
}
