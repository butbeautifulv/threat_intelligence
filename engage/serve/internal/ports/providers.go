package ports

import (
	"context"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cve"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/intelligence"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/workflow"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// IntelProvider is the intelligence surface used by HTTP and MCP handlers.
type IntelProvider interface {
	AnalyzeTarget(ctx context.Context, req contract.AnalyzeTargetRequest) contract.AnalyzeTargetResponse
	SelectTools(ctx context.Context, targetType, objective string) []string
	OptimizeParameters(targetType, toolName string, params map[string]string) map[string]string
	CreateAttackChain(ctx context.Context, target, objective string) map[string]any
	TechnologyDetection(ctx context.Context, target string) map[string]any
	ComprehensiveAPIAudit(ctx context.Context, subject string, req intelligence.ComprehensiveAPIAuditRequest) map[string]any
	CorrelateThreatIntelligence(ctx context.Context, target, indicators string) map[string]any
	TargetTimeline(ctx context.Context, req intelligence.TargetTimelineRequest) intelligence.TargetTimelineResponse
	DiscoverAttackChains(ctx context.Context, target, objective string) map[string]any
	ExecuteAttackChain(ctx context.Context, subject, target, objective string, parallel bool) map[string]any
	AIVulnerabilityAssessment(ctx context.Context, subject, target string, maxTools int) map[string]any
}

// CVEProvider is the CVE / vuln-intel surface used by HTTP and MCP handlers.
type CVEProvider interface {
	MonitorFromBody(ctx context.Context, body map[string]any) cve.MonitorResult
	GenerateExploitFromCVE(ctx context.Context, body map[string]any) cve.ExploitResult
	Lookup(ctx context.Context, cveID string) map[string]any
}

// CTFProvider is the CTF surface used by HTTP and MCP handlers.
type CTFProvider interface {
	CreateChallengeWorkflow(ch ctf.Challenge) (map[string]any, error)
	AutoSolve(ctx context.Context, subject string, ch ctf.Challenge, executeTools bool, maxSteps int) (map[string]any, error)
	SuggestTools(description, category, target string) map[string]any
	TeamStrategy(challenges []ctf.Challenge, teamSkills map[string][]string) map[string]any
	AnalyzeCrypto(text, cipherType, keyHint, knownPlaintext, additionalInfo string) map[string]any
	AnalyzeForensics(ctx context.Context, subject string, path string, opts ctf.ForensicsOptions) map[string]any
	AnalyzeBinary(ctx context.Context, subject string, path string, opts ctf.BinaryOptions) map[string]any
	RunPlaybook(ctx context.Context, subject string, pb workflow.Playbook, target string, execute bool) map[string]any
}
