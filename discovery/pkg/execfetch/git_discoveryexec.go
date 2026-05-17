//go:build discoveryexec

package execfetch

import (
	"context"
	"time"

	"github.com/butbeautifulv/veil/pkg/exec"
)

// GitClone runs a shallow git clone via pkg/exec (spike for discovery-fetcher profile).
func GitClone(ctx context.Context, workDir, repoURL string) exec.Result {
	ex := &exec.Executor{
		WorkDir: workDir,
		Sandbox: exec.NewSandboxFromEnv(),
	}
	return ex.Run(ctx, "git", []string{"clone", "--depth", "1", repoURL, "."}, 10*time.Minute, nil)
}
