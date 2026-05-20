package config

import (
	"strings"
	"testing"
)

func TestValidateExecutionProfile(t *testing.T) {
	cases := []struct {
		name      string
		profile   string
		mode      string
		container string
		wantErr   bool
	}{
		{name: "client_native_local", profile: ExecutionProfileClientNative, mode: "local", container: "", wantErr: false},
		{name: "client_native_empty_mode", profile: ExecutionProfileClientNative, mode: "", container: "", wantErr: false},
		{name: "client_native_docker", profile: ExecutionProfileClientNative, mode: "docker", container: "", wantErr: true},
		{name: "client_native_container", profile: ExecutionProfileClientNative, mode: "local", container: "engage-runner", wantErr: true},
		{name: "docker_exec_docker", profile: ExecutionProfileDockerExec, mode: "docker", container: "engage-runner", wantErr: false},
		{name: "unsupported_profile", profile: "k8s-exec", mode: "local", container: "", wantErr: true},
		{name: "client_native_mode_other", profile: ExecutionProfileClientNative, mode: "remote", container: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ENGAGE_RUNNER_MODE", tc.mode)
			t.Setenv("ENGAGE_RUNNER_CONTAINER", tc.container)
			cfg := &Config{ExecutionProfile: tc.profile}
			err := cfg.ValidateExecutionProfile()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateExecutionProfile_unsupportedMessage(t *testing.T) {
	t.Setenv("ENGAGE_RUNNER_MODE", "local")
	t.Setenv("ENGAGE_RUNNER_CONTAINER", "")
	cfg := &Config{ExecutionProfile: "bad"}
	err := cfg.ValidateExecutionProfile()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), ExecutionProfileClientNative) || !strings.Contains(err.Error(), ExecutionProfileDockerExec) {
		t.Fatalf("error should mention supported profiles: %v", err)
	}
}
