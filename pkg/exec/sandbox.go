package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// commandContext is overridden in tests to avoid real docker exec.
var commandContext = exec.CommandContext

// Sandbox runs tools inside an isolated container via docker exec.
type Sandbox struct {
	Mode      string // local | docker
	Container string
	WorkDir   string
}

func (s *Sandbox) Enabled() bool {
	return s != nil && strings.EqualFold(s.Mode, "docker") && s.Container != ""
}

func (s *Sandbox) Exec(ctx context.Context, binary string, args []string, timeout time.Duration, proc ProcessTracker, track *TrackInfo) Result {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	workDir := s.WorkDir
	if workDir == "" {
		workDir = "/tmp/engage"
	}
	quoted := make([]string, 0, len(args)+1)
	quoted = append(quoted, binary)
	for _, a := range args {
		quoted = append(quoted, shellQuote(a))
	}
	inner := strings.Join(quoted, " ")
	cmdArgs := []string{
		"exec", "-i", s.Container,
		"sh", "-c",
		fmt.Sprintf("cd %s && %s", shellQuote(workDir), inner),
	}
	cmd := commandContext(ctx, "docker", cmdArgs...)
	cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}

	var trackPID int
	if proc != nil && track != nil {
		session := s.Container + ":" + inner
		trackPID = proc.RegisterDocker(track.Tool, track.Target, "docker exec "+session, session)
	}
	res := runCmd(cmd)
	if proc != nil && track != nil {
		st := "done"
		if res.ExitCode != 0 || res.Err != nil {
			st = "failed"
		}
		out := res.Stdout + res.Stderr
		proc.UpdateProgress(trackPID, 1, out, int64(len(out)))
		proc.Finish(trackPID, st)
	}
	return res
}

func runCmd(cmd *exec.Cmd) Result {
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
			res.Err = err
		}
	}
	return res
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.IndexFunc(s, func(r rune) bool {
		return r == ' ' || r == '\'' || r == '"' || r == '$' || r == '`' || r == '\\'
	}) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// NewSandboxFromEnv builds sandbox config from environment.
func NewSandboxFromEnv() *Sandbox {
	mode := strings.TrimSpace(os.Getenv("ENGAGE_RUNNER_MODE"))
	if mode == "" {
		mode = "local"
	}
	return &Sandbox{
		Mode:      mode,
		Container: strings.TrimSpace(os.Getenv("ENGAGE_RUNNER_CONTAINER")),
		WorkDir:   strings.TrimSpace(os.Getenv("ENGAGE_RUNNER_WORKDIR")),
	}
}
