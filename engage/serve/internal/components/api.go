package components

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/auth"
)

type APIComponents struct {
	Auth      *auth.Stack
	Registry  *tools.Registry
	Tools     *toolsuc.Runner
	Intel     *intelligence.Service
	Workflows *workflow.Service
	Jobs      *jobuc.Queue
	Processes *process.Manager
	Audit     *audit.Logger
}

func InitAPI(cfg *config.Config, logger interface{ Info(string, ...any) }) (*APIComponents, error) {
	if err := config.ValidateSecurity(cfg.Security, cfg.Auth.Enabled); err != nil {
		return nil, err
	}
	wd, _ := os.Getwd()
	catalogPath := cfg.CatalogPath
	if !filepath.IsAbs(catalogPath) {
		catalogPath = filepath.Join(wd, catalogPath)
	}
	catalogDir := filepath.Dir(catalogPath)
	livePath := filepath.Join(catalogDir, "tools.live.yaml")
	enabledPath := filepath.Join(catalogDir, "tools.enabled.yaml")
	specs, err := tools.LoadCatalog(livePath, enabledPath, catalogPath)
	if err != nil {
		return nil, fmt.Errorf("catalog: %w", err)
	}
	stack, err := newAuthStack(context.Background(), cfg.Auth)
	if err != nil {
		return nil, err
	}
	reg := tools.NewRegistry(specs)
	exec := &runner.Executor{
		WorkDir: cfg.RunnerWork,
		Sandbox: runner.NewSandboxFromEnv(),
	}
	_ = os.MkdirAll(cfg.RunnerWork, 0o700)
	auditLog := audit.New(nil)
	toolRunner := &toolsuc.Runner{
		Registry: reg,
		Exec:     exec,
		Audit:    auditLog,
		Auth:     stack,
	}
	veil := veilgraph.New(veilgraph.Config{
		BaseURL:      cfg.VeilAPI.BaseURL,
		ClientID:     cfg.VeilAPI.ClientID,
		ClientSecret: cfg.VeilAPI.ClientSecret,
		TokenURL:     cfg.VeilAPI.TokenURL,
	})
	intel := &intelligence.Service{Veil: veil, Registry: reg}
	wf := &workflow.Service{Intel: intel, Tools: toolRunner}
	jobs := jobuc.NewQueue(toolRunner, 2)
	return &APIComponents{
		Auth:      stack,
		Registry:  reg,
		Tools:     toolRunner,
		Intel:     intel,
		Workflows: wf,
		Jobs:      jobs,
		Processes: process.NewManager(),
		Audit:     auditLog,
	}, nil
}
