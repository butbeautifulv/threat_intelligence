package components

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
	"github.com/butbeautifulv/veil/engage/serve/internal/config"
	"github.com/butbeautifulv/veil/engage/serve/internal/ports"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/security"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cache"
	cmduc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/command"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/recovery"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/files"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/browser"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/bugbounty"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/visual"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	jobuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/job"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/process"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tooldispatch"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/auth"
)

// Provider aliases for HTTP/MCP facades (definitions in ports avoid import cycles with mcpserver).
type (
	IntelProvider = ports.IntelProvider
	CVEProvider   = ports.CVEProvider
	CTFProvider   = ports.CTFProvider
)

type APIComponents struct {
	Auth      *auth.Stack
	Registry  *tools.Registry
	Tools     *toolsuc.Runner
	ToolDispatch *tooldispatch.Dispatcher
	Intel     IntelProvider
	CVE       CVEProvider
	CTF        CTFProvider
	BugBounty  *bugbounty.Service
	Browser    *browser.Service
	Workflows  *workflow.Service
	Progress   *visual.Store
	Jobs      *jobuc.Queue
	Processes *process.Manager
	Audit       *audit.Logger
	AuditStore  *audit.Store
	AuditReader audit.Reader
	Files      *files.Manager
	Cache     *cache.Store
	Command            *cmduc.Runner
	StartedAt          time.Time
	MetricsEnabled     bool
	AuditWebhookURL    string
	AuditWebhookSecret string
	CatalogPath          string
}

func InitAPI(cfg *config.Config, logger interface{ Info(string, ...any) }) (*APIComponents, error) {
	if err := config.ValidateSecurity(cfg.Security, cfg.Auth.Enabled); err != nil {
		return nil, err
	}
	if err := cfg.ValidateExecutionProfile(); err != nil {
		return nil, err
	}
	if findings := security.RunSelfTest(cfg.Security, cfg.Auth.Enabled); len(findings) > 0 {
		logger.Info(security.FormatReport(findings))
		if os.Getenv("ENGAGE_HARDENING_FAIL_ON") == "high" {
			if err := security.FailOn(findings, security.SeverityHigh); err != nil {
				return nil, err
			}
		}
	}
	wd, _ := os.Getwd()
	catalogPath := cfg.CatalogPath
	if !filepath.IsAbs(catalogPath) {
		catalogPath = filepath.Join(wd, catalogPath)
	}
	catalogDir := filepath.Dir(catalogPath)
	livePath := filepath.Join(catalogDir, "tools.live.yaml")
	enabledPath := filepath.Join(catalogDir, "tools.enabled.yaml")
	// Base catalog first; tools.live.yaml and tools.enabled.yaml override (enable runner tools).
	specs, err := tools.LoadCatalog(catalogPath, livePath, enabledPath)
	if err != nil {
		return nil, fmt.Errorf("catalog: %w", err)
	}
	stack, err := newAuthStack(context.Background(), cfg.Auth)
	if err != nil {
		return nil, err
	}
	reg := tools.NewRegistry(specs)
	procMgr := process.NewManager()
	exec := &runner.Executor{
		WorkDir:   cfg.RunnerWork,
		Sandbox:   runner.NewSandboxFromEnv(),
		Processes: procMgr,
	}
	_ = os.MkdirAll(cfg.RunnerWork, 0o700)
	var auditStore *audit.Store
	var auditPG *audit.PostgresStore
	var auditAppenders []audit.Appender
	if cfg.AuditDir != "" {
		auditStore, _ = audit.NewStore(cfg.AuditDir)
		if auditStore != nil {
			auditAppenders = append(auditAppenders, auditStore)
		}
	}
	if cfg.AuditPostgresURL != "" {
		if pg, err := audit.NewPostgresStore(cfg.AuditPostgresURL); err == nil {
			auditPG = pg
			auditAppenders = append(auditAppenders, pg)
			if cfg.AuditRetentionDays > 0 {
				_, _ = pg.Retention(time.Duration(cfg.AuditRetentionDays) * 24 * time.Hour)
			}
		}
	}
	var auditReader audit.Reader
	if auditPG != nil {
		auditReader = auditPG
	} else {
		auditReader = auditStore
	}
	var auditAppender audit.Appender
	if len(auditAppenders) > 0 {
		auditAppender = audit.NewMultiStore(auditAppenders...)
	}
	auditLog := audit.NewWithStore(nil, auditAppender)
	var eventPub *engageevents.Publisher
	if cfg.EventsNATSEnabled && cfg.NATSURL != "" {
		if pub, err := engageevents.Connect(cfg.NATSURL, cfg.EventsNATSSubject); err == nil {
			eventPub = pub
			auditLog.SetEventPublisher(pub)
		} else if logger != nil {
			logger.Info("engage events NATS connect failed", "url", cfg.NATSURL, "err", err.Error())
		}
	}
	resultCache := cache.New(15 * time.Minute)
	toolRunner := &toolsuc.Runner{
		Registry:    reg,
		Exec:        exec,
		Audit:       auditLog,
		Auth:        stack,
		Cache:       resultCache,
		Recovery:    recovery.Default(),
		TargetGuard: security.ParseTargetGuardMode(os.Getenv),
	}
	veil := veilgraph.New(veilgraph.Config{
		BaseURL:      cfg.VeilAPI.BaseURL,
		ClientID:     cfg.VeilAPI.ClientID,
		ClientSecret: cfg.VeilAPI.ClientSecret,
		TokenURL:     cfg.VeilAPI.TokenURL,
	})
	cveSvc := cve.NewService(veil, cve.DefaultNVDClient())
	intel := &intelligence.Service{
		Veil:     veil,
		Registry: reg,
		Engine:   intelligence.DefaultDecisionEngine(),
		Tools:    toolRunner,
		Audit:    auditReader,
		CVE:      cveSvc,
	}
	var jobStore jobuc.Store = jobuc.NewMemoryStore()
	switch cfg.JobsMode {
	case jobuc.ModeFile:
		_ = os.MkdirAll(cfg.JobsDir, 0o700)
		jobStore = jobuc.NewFileStore(cfg.JobsDir)
	case jobuc.ModeRedis:
		rs, err := jobuc.NewRedisStore(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("redis jobs: %w", err)
		}
		jobStore = rs
	case jobuc.ModeNats:
		ns, err := jobuc.NewNATSStore(cfg.NATSURL)
		if err != nil {
			return nil, fmt.Errorf("nats jobs: %w", err)
		}
		jobStore = ns
	}
	jobs := jobuc.NewQueue(toolRunner,
		jobuc.WithStore(jobStore),
		jobuc.WithMode(cfg.JobsMode),
		jobuc.WithPollInterval(cfg.JobsPollInterval),
		jobuc.WithConcurrency(cfg.WorkerConcurrency),
	)
	progressStore := visual.NewStore()
	browserSvc := browser.NewServiceFromEnv()
	wf := &workflow.Service{
		Intel: intel, Tools: toolRunner, Jobs: jobs, Progress: progressStore,
		MaxParallel: cfg.MaxParallel,
	}
	intel.ParallelRunner = wf
	if eventPub != nil {
		wf.Findings = findingBridge{pub: eventPub}
	}
	bbSvc := bugbounty.NewService(reg, intel, wf, wf.Findings, jobs)
	wf.BugBounty = bbSvc
	fileMgr, err := files.NewManager(cfg.FilesDir)
	if err != nil {
		return nil, fmt.Errorf("files: %w", err)
	}
	cmdRunner := cmduc.New(exec, reg, cfg.Security.AllowRawCommand)
	ctfSvc := ctf.NewService(reg, toolRunner, intel, cfg.FilesDir)
	toolDispatch := tooldispatch.NewDispatcher(toolRunner, intel, cveSvc, ctfSvc, bbSvc, browserSvc, procMgr, wf, catalogPath, fileMgr)
	return &APIComponents{
		Auth:               stack,
		Registry:           reg,
		Tools:              toolRunner,
		ToolDispatch:       toolDispatch,
		Intel:              intel,
		CVE:                cveSvc,
		CTF:                ctfSvc,
		BugBounty:          bbSvc,
		Browser:            browserSvc,
		Workflows:          wf,
		Progress:           progressStore,
		Jobs:               jobs,
		Processes:          procMgr,
		Audit:              auditLog,
		AuditStore:         auditStore,
		AuditReader:        auditReader,
		Files:              fileMgr,
		Cache:              resultCache,
		Command:            cmdRunner,
		StartedAt:          time.Now().UTC(),
		MetricsEnabled:     cfg.MetricsEnabled,
		AuditWebhookURL:    cfg.AuditWebhookURL,
		AuditWebhookSecret: cfg.AuditWebhookSecret,
		CatalogPath:        catalogPath,
	}, nil
}

var (
	_ IntelProvider = (*intelligence.Service)(nil)
	_ CVEProvider   = (*cve.Service)(nil)
	_ CTFProvider   = (*ctf.Service)(nil)
)
