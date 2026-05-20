package exec

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestLookupBinary(t *testing.T) {
	path, err := LookupBinary("echo")
	if err != nil || path == "" {
		t.Fatalf("echo: path=%q err=%v", path, err)
	}
	if _, err := LookupBinary("definitely-not-a-binary-xyz"); err == nil {
		t.Fatal("expected missing binary error")
	}
}

func TestSubstitutePlaceholders_unresolved(t *testing.T) {
	if got := substitutePlaceholders("{missing}", map[string]string{"other": "x"}); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := substitutePlaceholders("ok", map[string]string{}); got != "ok" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildArgs_edgeCases(t *testing.T) {
	if got := BuildArgs([]string{"", "{target}"}, "t", "", nil); len(got) != 1 || got[0] != "t" {
		t.Fatalf("skip empty: %v", got)
	}
	got := BuildArgs([]string{"{additional_args}", "x"}, "t", "-v", nil)
	if !contains(got, "-v") || !contains(got, "x") {
		t.Fatalf("additional_args: %v", got)
	}
	got = BuildArgs([]string{"-p", "{ports}"}, "t", "", map[string]string{"ports": ""})
	for _, a := range got {
		if a == "-p" {
			t.Fatalf("skip flag pair: %v", got)
		}
	}
	got = BuildArgs([]string{"{unknown}"}, "t", "", nil)
	if len(got) != 0 {
		t.Fatalf("unresolved placeholder: %v", got)
	}
	got = BuildArgs([]string{"-{additional_args}"}, "t", "a", nil)
	if !contains(got, "-") || !contains(got, "a") {
		t.Fatalf("additional prefix strip: %v", got)
	}
}

func TestFilterEnv_strictAndEmpty(t *testing.T) {
	t.Setenv("ENGAGE_STRICT_ENV", "1")
	t.Setenv("ENGAGE_ENV", "prod")
	out := filterEnv([]string{"HOME=/h", "PATH=/bin", "SECRET=x"})
	for _, e := range out {
		if strings.HasPrefix(e, "HOME=") {
			t.Fatalf("HOME should be stripped in strict: %v", out)
		}
	}
	out = filterEnv(nil)
	if len(out) == 0 || !strings.HasPrefix(out[0], "PATH=") {
		t.Fatalf("default PATH: %v", out)
	}
}

func TestMergeEngagePathExtra_noPATH(t *testing.T) {
	t.Setenv("ENGAGE_PATH_EXTRA", "/extra/bin")
	out := mergeEngagePathExtra([]string{"LANG=C"})
	found := false
	for _, e := range out {
		if strings.HasPrefix(e, "PATH=/extra/bin:") {
			found = true
		}
	}
	if !found {
		t.Fatalf("PATH append: %v", out)
	}
}

func TestExecutor_sandboxRun(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "echo", "sandbox")
	}
	e := &Executor{Sandbox: &Sandbox{Mode: "docker", Container: "c"}}
	res := e.Run(context.Background(), "tool", nil, time.Second, nil)
	if res.ExitCode != 0 || !strings.Contains(res.Stdout, "sandbox") {
		t.Fatalf("res: %+v", res)
	}
}

func TestExecutor_runLocalWithTracker(t *testing.T) {
	proc := &mockTracker{}
	e := &Executor{Processes: proc}
	res := e.Run(context.Background(), "echo", []string{"tracked"}, time.Second, &TrackInfo{Tool: "t", Target: "h"})
	if res.ExitCode != 0 || proc.finished != 1 {
		t.Fatalf("res=%+v finished=%d", res, proc.finished)
	}
}

func TestRunLocal_startErrorAndNonExitError(t *testing.T) {
	res := runLocal(context.Background(), "", "/no/such/binary", nil, time.Second, nil, nil)
	if res.ExitCode != -1 || res.Err == nil {
		t.Fatalf("start fail: %+v", res)
	}
	old := commandContext
	defer func() { commandContext = old }()
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		c := exec.CommandContext(ctx, "false")
		c.Path = "/bin/false"
		return c
	}
	// runLocal uses exec.CommandContext directly, not commandContext — use invalid command
	res = runLocal(context.Background(), "", "\x00", nil, time.Second, nil, nil)
	if res.ExitCode != -1 {
		t.Fatalf("invalid binary: %+v", res)
	}
}

func TestSandbox_execDefaultsAndFailure(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "exit 3")
	}
	s := &Sandbox{Mode: "docker", Container: "c", WorkDir: ""}
	res := s.Exec(context.Background(), "x", nil, 0, nil, nil)
	if res.ExitCode != 3 {
		t.Fatalf("exit: %+v", res)
	}
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		c := exec.CommandContext(ctx, "sh", "-c", "echo ok")
		c.Path = "/nonexistent-sh"
		return c
	}
	res = runCmd(commandContext(context.Background(), "sh"))
	if res.ExitCode != -1 || res.Err == nil {
		t.Fatalf("runCmd err: %+v", res)
	}
}

func TestSandbox_execTrackerFailed(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "exit 1")
	}
	proc := &mockTracker{}
	s := &Sandbox{Mode: "docker", Container: "c"}
	_ = s.Exec(context.Background(), "t", nil, time.Second, proc, &TrackInfo{Tool: "t", Target: "h"})
	if proc.finished != 1 {
		t.Fatalf("finished %d", proc.finished)
	}
}

func TestExecutor_sandboxEnabledNil(t *testing.T) {
	var e *Executor
	if e.sandboxEnabled() {
		t.Fatal("nil executor")
	}
	e = &Executor{}
	if e.sandboxEnabled() {
		t.Fatal("nil sandbox")
	}
}

func TestExecutionProfileClientNative(t *testing.T) {
	t.Setenv("ENGAGE_EXECUTION_PROFILE", "client-native")
	if !executionProfileClientNative() {
		t.Fatal("expected true")
	}
	os.Unsetenv("ENGAGE_EXECUTION_PROFILE")
	if executionProfileClientNative() {
		t.Fatal("expected false")
	}
}

func TestRunLocal_exitError(t *testing.T) {
	res := runLocal(context.Background(), "", "sh", []string{"-c", "exit 2"}, time.Second, nil, nil)
	if res.ExitCode != 2 {
		t.Fatalf("exit: %+v", res)
	}
}

func TestRunLocal_zeroTimeout(t *testing.T) {
	res := runLocal(context.Background(), "", "echo", []string{"x"}, 0, nil, nil)
	if res.ExitCode != 0 {
		t.Fatalf("res: %+v", res)
	}
}

func TestRunLocal_trackerStartFail(t *testing.T) {
	proc := &mockTracker{}
	res := runLocal(context.Background(), "", "\x00invalid", nil, time.Second, proc, &TrackInfo{Tool: "t"})
	if res.ExitCode != -1 {
		t.Fatalf("res: %+v", res)
	}
}

func TestRunLocal_trackerCommandFail(t *testing.T) {
	proc := &mockTracker{}
	res := runLocal(context.Background(), "", "sh", []string{"-c", "exit 5"}, time.Second, proc, &TrackInfo{Tool: "t", Target: "h"})
	if res.ExitCode != 5 || proc.finished != 1 {
		t.Fatalf("res=%+v finished=%d", res, proc.finished)
	}
}

// Ensure mockTracker satisfies ProcessTracker (compile-time).
var _ ProcessTracker = (*mockTracker)(nil)

func TestRunCmd_success(t *testing.T) {
	res := runCmd(exec.Command("echo", "ok"))
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("res: %+v", res)
	}
}

