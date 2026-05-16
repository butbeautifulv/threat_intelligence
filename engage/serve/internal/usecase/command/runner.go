package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cache"
)

// Runner executes guarded shell commands (catalog binary allowlist or lab raw mode).
type Runner struct {
	Exec           *runner.Executor
	Registry       *tools.Registry
	AllowRaw       bool
	AllowedBinaries map[string]struct{}
}

func New(exec *runner.Executor, reg *tools.Registry) *Runner {
	allowed := make(map[string]struct{})
	if reg != nil {
		for _, s := range reg.List() {
			if s.Binary != "" {
				allowed[s.Binary] = struct{}{}
			}
		}
	}
	return &Runner{
		Exec:            exec,
		Registry:        reg,
		AllowRaw:        os.Getenv("ENGAGE_ALLOW_RAW_COMMAND") == "1",
		AllowedBinaries: allowed,
	}
}

func (r *Runner) Run(ctx context.Context, command string, useCache bool, c *cache.Store) map[string]any {
	command = strings.TrimSpace(command)
	if command == "" {
		return map[string]any{"success": false, "error": "command is required"}
	}
	if useCache && c != nil {
		if v, ok := c.Get(command); ok {
			return map[string]any{"success": true, "output": v, "cached": true}
		}
	}
	bin, args, err := r.parseCommand(command)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	res := r.Exec.Run(ctx, bin, args, 5*time.Minute, &runner.TrackInfo{Tool: "command", Target: bin})
	out := res.Stdout
	if res.Stderr != "" {
		out = strings.TrimSpace(out + "\n" + res.Stderr)
	}
	ok := res.Err == nil && res.ExitCode == 0
	if useCache && ok && c != nil {
		c.Set(command, out)
	}
	errMsg := ""
	if res.Err != nil {
		errMsg = res.Err.Error()
	} else if res.ExitCode != 0 {
		errMsg = fmt.Sprintf("exit code %d", res.ExitCode)
	}
	return map[string]any{
		"success":   ok,
		"output":    out,
		"error":     errMsg,
		"exit_code": res.ExitCode,
		"command":   command,
	}
}

func (r *Runner) parseCommand(command string) (string, []string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	bin := parts[0]
	args := parts[1:]
	if r.AllowRaw {
		return bin, args, nil
	}
	if _, ok := r.AllowedBinaries[bin]; ok {
		return bin, args, nil
	}
	return "", nil, fmt.Errorf("binary %q not in catalog allowlist (set ENGAGE_ALLOW_RAW_COMMAND=1 for lab)", bin)
}
