package runner

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
	WorkDir string
	Sandbox *Sandbox
}

func (e *Executor) Run(ctx context.Context, binary string, args []string, timeout time.Duration) Result {
	if e.Sandbox != nil && e.Sandbox.Enabled() {
		return e.Sandbox.Exec(ctx, binary, args, timeout)
	}
	return runLocal(ctx, e.WorkDir, binary, args, timeout)
}

func runLocal(ctx context.Context, workDir, binary string, args []string, timeout time.Duration) Result {
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

	err := cmd.Run()
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
	allow := map[string]bool{
		"PATH": true, "HOME": true, "USER": true, "LANG": true, "LC_ALL": true,
		"TMPDIR": true, "TMP": true, "TEMP": true,
	}
	var out []string
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		if allow[k] || strings.HasPrefix(k, "ENGAGE_") {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		return []string{"PATH=/usr/local/bin:/usr/bin:/bin"}
	}
	return out
}
