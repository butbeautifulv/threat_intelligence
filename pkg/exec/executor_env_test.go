package exec

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFilterEnv_engagesPathExtra(t *testing.T) {
	t.Setenv("ENGAGE_PATH_EXTRA", "/opt/veil-tools/bin:/extra")
	t.Setenv("ENGAGE_ENV", "local")
	t.Setenv("ENGAGE_STRICT_ENV", "0")
	out := filterEnv([]string{"PATH=/usr/bin:/bin", "FOO=secret"})
	var pathLine string
	for _, e := range out {
		if strings.HasPrefix(e, "PATH=") {
			pathLine = e
			break
		}
	}
	if pathLine == "" {
		t.Fatal("expected PATH in filtered env")
	}
	if !strings.HasPrefix(pathLine, "PATH=/opt/veil-tools/bin:/extra:") {
		t.Fatalf("PATH should start with ENGAGE_PATH_EXTRA dirs: %s", pathLine)
	}
}

func TestExecutorRun_clientNativeSkipsDockerSandbox(t *testing.T) {
	t.Setenv("ENGAGE_EXECUTION_PROFILE", "client-native")
	t.Setenv("ENGAGE_RUNNER_MODE", "local")
	e := &Executor{
		WorkDir: t.TempDir(),
		Sandbox: &Sandbox{Mode: "docker", Container: "fake"},
	}
	res := e.Run(context.Background(), "/bin/true", []string{}, 2*time.Second, nil)
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("expected local /bin/true; got code=%d err=%v stderr=%q", res.ExitCode, res.Err, res.Stderr)
	}
}
