package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/pkg/auth"
)

type Config struct {
	ListenAddr  string
	Env         string
	Auth        auth.Config
	Security    SecurityConfig
	CatalogPath      string
	RunnerWork       string
	FilesDir         string
	AuditDir         string
	JobsMode         string
	JobsDir          string
	RedisURL         string
	NATSURL          string
	JobsPollInterval    time.Duration
	WorkerConcurrency   int
	MaxParallel         int
	VeilAPI             VeilAPIConfig
	MCPHTTP          MCPHTTPConfig
	MetricsEnabled       bool
	AuditWebhookURL      string
	AuditWebhookSecret   string
	AuditPostgresURL     string
	AuditRetentionDays   int
	EventsNATSEnabled    bool
	EventsNATSSubject    string
}

type VeilAPIConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	TokenURL     string
}

type MCPHTTPConfig struct {
	Enabled   bool
	Listen    string
	Path      string
	BindLocal bool
}

func LoadAPI() *Config {
	return loadBase(getenv("ENGAGE_API_LISTEN", ":8890"), getenv("ENGAGE_ENV", "local"))
}

func LoadMCP() *Config {
	return loadBase("", getenv("ENGAGE_ENV", "local"))
}

func loadBase(listen, env string) *Config {
	return &Config{
		ListenAddr:  listen,
		Env:         env,
		Auth:        loadAuthFromEnv(),
		Security:    LoadSecurityForEnv(env),
		CatalogPath:      getenv("ENGAGE_CATALOG_PATH", "catalog/tools.yaml"),
		RunnerWork:       getenv("ENGAGE_RUNNER_WORKDIR", "/tmp/engage"),
		FilesDir:         getenv("ENGAGE_FILES_DIR", "/var/veil/engage/files"),
		AuditDir:         getenv("ENGAGE_AUDIT_DIR", "/var/veil/engage/audit"),
		JobsMode:         getenv("ENGAGE_JOBS_MODE", "memory"),
		JobsDir:          getenv("ENGAGE_JOBS_DIR", "/tmp/engage/jobs"),
		RedisURL:         getenv("ENGAGE_REDIS_URL", "redis://127.0.0.1:6379/0"),
		NATSURL:          getenv("ENGAGE_NATS_URL", "nats://127.0.0.1:4222"),
		JobsPollInterval:  jobsPollInterval(),
		WorkerConcurrency: workerConcurrency(),
		MaxParallel:       maxParallel(),
		VeilAPI: VeilAPIConfig{
			BaseURL:      getenv("ENGAGE_VEIL_API_URL", "http://localhost:8090"),
			ClientID:     getenv("ENGAGE_VEIL_CLIENT_ID", ""),
			ClientSecret: getenv("ENGAGE_VEIL_CLIENT_SECRET", ""),
			TokenURL:     getenv("ENGAGE_VEIL_TOKEN_URL", ""),
		},
		MCPHTTP: MCPHTTPConfig{
			Enabled:   envBool("ENGAGE_MCP_HTTP_ENABLED", false),
			Listen:    getenv("ENGAGE_MCP_HTTP_LISTEN", ":8892"),
			Path:      getenv("ENGAGE_MCP_HTTP_PATH", "/mcp"),
			BindLocal: envBool("ENGAGE_MCP_HTTP_BIND_LOCAL", false),
		},
		MetricsEnabled:     envBool("ENGAGE_METRICS_ENABLED", false),
		AuditWebhookURL:    getenv("ENGAGE_AUDIT_WEBHOOK_URL", ""),
		AuditWebhookSecret: getenv("ENGAGE_AUDIT_WEBHOOK_SECRET", ""),
		AuditPostgresURL:   getenv("ENGAGE_AUDIT_POSTGRES_URL", ""),
		AuditRetentionDays: auditRetentionDays(),
		EventsNATSEnabled:  envBool("ENGAGE_EVENTS_NATS_ENABLED", false),
		EventsNATSSubject:  getenv("ENGAGE_EVENTS_NATS_SUBJECT", "engage.events.audit"),
	}
}

func auditRetentionDays() int {
	if v := strings.TrimSpace(os.Getenv("ENGAGE_AUDIT_RETENTION_DAYS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func loadAuthFromEnv() auth.Config {
	c := auth.LoadConfigFromEnv()
	// Engage defaults when using engage-specific env prefix
	if v := getenv("ENGAGE_AUTH_ENABLED", ""); v != "" {
		c.Enabled = envBool("ENGAGE_AUTH_ENABLED", c.Enabled)
	}
	return c
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func workerConcurrency() int {
	n := 2
	if v := strings.TrimSpace(os.Getenv("ENGAGE_WORKER_CONCURRENCY")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	return n
}

func maxParallel() int {
	n := 5
	if v := strings.TrimSpace(os.Getenv("ENGAGE_MAX_PARALLEL")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	if n > 32 {
		n = 32
	}
	if n < 1 {
		n = 1
	}
	return n
}

func jobsPollInterval() time.Duration {
	sec := 1
	if v := strings.TrimSpace(os.Getenv("ENGAGE_JOBS_POLL_SEC")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sec = n
		}
	}
	return time.Duration(sec) * time.Second
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
