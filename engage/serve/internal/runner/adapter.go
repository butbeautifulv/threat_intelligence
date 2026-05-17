package runner

import (
	"context"
	"time"

	"github.com/butbeautifulv/veil/pkg/exec"
)

// Re-export exec primitives for engage callers.
type (
	Result         = exec.Result
	TrackInfo      = exec.TrackInfo
	ProcessTracker = exec.ProcessTracker
	Sandbox        = exec.Sandbox
)

var (
	BuildArgs         = exec.BuildArgs
	NewSandboxFromEnv = exec.NewSandboxFromEnv
)

// Executor runs allowlisted binaries; browser catalog tools use the sidecar when configured.
type Executor struct {
	WorkDir   string
	Sandbox   *Sandbox
	Processes ProcessTracker
}

func (e *Executor) exec() *exec.Executor {
	return &exec.Executor{
		WorkDir:   e.WorkDir,
		Sandbox:   e.Sandbox,
		Processes: e.Processes,
	}
}

func (e *Executor) Run(ctx context.Context, binary string, args []string, timeout time.Duration, track *TrackInfo) Result {
	if proxy := NewBrowserProxyFromEnv(); proxy != nil && proxy.Enabled() && IsBrowserBinary(binary) {
		if timeout <= 0 {
			timeout = 5 * time.Minute
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		return proxy.Exec(ctx, target, args)
	}
	return e.exec().Run(ctx, binary, args, timeout, track)
}
