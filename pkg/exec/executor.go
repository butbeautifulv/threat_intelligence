package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Result holds subprocess output.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Executor runs allowlisted binaries with timeout.
type Executor struct {
	WorkDir   string
	Sandbox   *Sandbox
	Processes ProcessTracker
}

func (e *Executor) Run(ctx context.Context, binary string, args []string, timeout time.Duration, track *TrackInfo) Result {
	// Defense in depth: client-native execution must never use docker exec even if Sandbox is mis-wired.
	if e.sandboxEnabled() {
		return e.Sandbox.Exec(ctx, binary, args, timeout, e.Processes, track)
	}
	return runLocal(ctx, e.WorkDir, binary, args, timeout, e.Processes, track)
}

func (e *Executor) sandboxEnabled() bool {
	if e == nil || e.Sandbox == nil || !e.Sandbox.Enabled() {
		return false
	}
	if executionProfileClientNative() {
		return false
	}
	return true
}

func executionProfileClientNative() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ENGAGE_EXECUTION_PROFILE")), "client-native")
}

func runLocal(ctx context.Context, workDir, binary string, args []string, timeout time.Duration, proc ProcessTracker, track *TrackInfo) Result {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = filterEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var err error
	if proc != nil && track != nil {
		if err = cmd.Start(); err != nil {
			return Result{ExitCode: -1, Err: err}
		}
		pid := cmd.Process.Pid
		proc.Register(pid, track.Tool, track.Target, cmd.String())
		proc.UpdateProgress(pid, 0.05, "", 0)
		err = cmd.Wait()
		st := "done"
		if err != nil {
			st = "failed"
		}
		out := stdout.String() + stderr.String()
		proc.UpdateProgress(pid, 1, out, int64(len(out)))
		proc.Finish(pid, st)
	} else {
		err = cmd.Run()
	}

	res := Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
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

// BuildArgs substitutes {placeholders} from target, additional_args, and parameters.
func BuildArgs(template []string, target, additional string, parameters map[string]string) []string {
	vals := make(map[string]string, len(parameters)+2)
	for k, v := range parameters {
		vals[k] = v
	}
	vals["target"] = target
	vals["additional_args"] = additional

	out := make([]string, 0, len(template)+4)
	for i := 0; i < len(template); i++ {
		t := template[i]
		if t == "" {
			continue
		}
		if t == "{additional_args}" || strings.Contains(t, "{additional_args}") {
			t = strings.ReplaceAll(t, "{additional_args}", "")
			t = strings.TrimSpace(t)
			if t != "" {
				out = append(out, t)
			}
			if additional != "" {
				out = append(out, strings.Fields(additional)...)
			}
			continue
		}
		if strings.HasPrefix(t, "-") && i+1 < len(template) && strings.Contains(template[i+1], "{") {
			val := substitutePlaceholders(template[i+1], vals)
			if val == "" {
				i++
				continue
			}
			out = append(out, t, val)
			i++
			continue
		}
		expanded := substitutePlaceholders(t, vals)
		if expanded == "" {
			continue
		}
		out = append(out, strings.Fields(expanded)...)
	}
	return out
}

func substitutePlaceholders(s string, vals map[string]string) string {
	out := s
	for k, v := range vals {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	if strings.Contains(out, "{") {
		return ""
	}
	return strings.TrimSpace(out)
}

func LookupBinary(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("binary %q not found on PATH", name)
	}
	return path, nil
}

func filterEnv(env []string) []string {
	strict := os.Getenv("ENGAGE_STRICT_ENV") == "1" ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("ENGAGE_ENV")), "prod")
	allow := map[string]bool{
		"PATH": true, "LANG": true, "LC_ALL": true,
		"TMPDIR": true, "TMP": true, "TEMP": true,
	}
	if !strict {
		allow["HOME"] = true
		allow["USER"] = true
	}
	var out []string
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		if allow[k] || strings.HasPrefix(k, "ENGAGE_") {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		out = []string{"PATH=/usr/local/bin:/usr/bin:/bin"}
	}
	return mergeEngagePathExtra(out)
}

// mergeEngagePathExtra prepends ENGAGE_PATH_EXTRA (colon-separated dirs) to PATH when set.
func mergeEngagePathExtra(out []string) []string {
	extra := strings.TrimSpace(os.Getenv("ENGAGE_PATH_EXTRA"))
	if extra == "" {
		return out
	}
	for i, e := range out {
		k, v, ok := strings.Cut(e, "=")
		if !ok || k != "PATH" {
			continue
		}
		out[i] = "PATH=" + extra + ":" + v
		return out
	}
	return append(out, "PATH="+extra+":/usr/local/bin:/usr/bin:/bin")
}
