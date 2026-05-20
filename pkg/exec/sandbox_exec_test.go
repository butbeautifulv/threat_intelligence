package exec

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestShellQuote(t *testing.T) {
	if shellQuote("") != "''" {
		t.Fatal("empty")
	}
	if shellQuote("plain") != "plain" {
		t.Fatal("plain")
	}
	if shellQuote("has space") != "'has space'" {
		t.Fatalf("got %q", shellQuote("has space"))
	}
}

func TestRunCmd_successAndExitError(t *testing.T) {
	res := runCmd(exec.Command("echo", "hello"))
	if res.ExitCode != 0 || res.Stdout != "hello\n" {
		t.Fatalf("res: %+v", res)
	}
	resFail := runCmd(exec.Command("sh", "-c", "exit 7"))
	if resFail.ExitCode != 7 {
		t.Fatalf("exit: %+v", resFail)
	}
}

func TestSandbox_Exec_mockDocker(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		if name != "docker" {
			t.Fatalf("name %s", name)
		}
		return exec.CommandContext(ctx, "echo", "ok")
	}
	s := &Sandbox{Mode: "docker", Container: "runner", WorkDir: "/tmp/engage"}
	res := s.Exec(context.Background(), "nmap", []string{"-h"}, 5*time.Second, nil, nil)
	if res.ExitCode != 0 || res.Stdout != "ok\n" {
		t.Fatalf("res: %+v", res)
	}
}

func TestSandbox_Exec_withTracker(t *testing.T) {
	old := commandContext
	defer func() { commandContext = old }()
	commandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "echo", "tracked")
	}
	proc := &mockTracker{}
	s := &Sandbox{Mode: "docker", Container: "c"}
	_ = s.Exec(context.Background(), "tool", nil, time.Second, proc, &TrackInfo{Tool: "nmap", Target: "h"})
	if proc.finished != 1 {
		t.Fatalf("finished %d", proc.finished)
	}
}

type mockTracker struct {
	finished int
}

func (m *mockTracker) Register(pid int, tool, target, command string) {}
func (m *mockTracker) RegisterDocker(tool, target, cmd, session string) int { return 42 }
func (m *mockTracker) UpdateProgress(pid int, pct float64, out string, n int64) {}
func (m *mockTracker) Finish(pid int, status string) { m.finished++ }
