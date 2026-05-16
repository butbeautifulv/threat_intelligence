package intelligence

import (
	"context"
	"net/url"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// Service provides target analysis and tool selection.
type Service struct {
	Veil     *veilgraph.Client
	Registry *tools.Registry
}

func (s *Service) AnalyzeTarget(ctx context.Context, req contract.AnalyzeTargetRequest) contract.AnalyzeTargetResponse {
	target := strings.TrimSpace(req.Target)
	resp := contract.AnalyzeTargetResponse{
		Target:     target,
		TargetType: "unknown",
		RiskLevel:  "medium",
		Confidence: 0.5,
		Metadata:   map[string]any{},
	}
	if u, err := url.Parse(target); err == nil && u.Host != "" {
		if strings.Contains(u.Path, "/api") {
			resp.TargetType = "api"
		} else {
			resp.TargetType = "web"
		}
		resp.Technologies = []string{"http"}
	} else if strings.Count(target, ".") >= 3 {
		resp.TargetType = "ip"
	}
	if s.Veil != nil && s.Veil.Enabled() {
		if raw, err := s.Veil.Categories(ctx); err == nil {
			resp.Metadata["veil_categories"] = raw
			resp.Confidence = 0.7
		}
	}
	return resp
}

func (s *Service) SelectTools(_ context.Context, targetType string, _ string) []string {
	var short []string
	switch targetType {
	case "web", "api":
		short = []string{"httpx", "nuclei", "gobuster", "nikto"}
	case "ip":
		short = []string{"nmap", "rustscan", "nuclei"}
	default:
		short = []string{"nmap", "httpx", "subfinder"}
	}
	return tools.ResolveCatalogNames(short, s.Registry)
}

// OptimizeParameters suggests CLI flags for a tool against a target profile.
func (s *Service) OptimizeParameters(targetType, toolName string, params map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range params {
		out[k] = v
	}
	engine := DefaultDecisionEngine()
	toolID := toolName
	if s.Registry != nil {
		toolID = tools.ResolveCatalogName(toolName, s.Registry)
	}
	if spec, ok := s.Registry.Get(toolID); ok && spec.Binary != "" {
		toolID = spec.Binary
	}
	optimized := engine.OptimizeParameters(targetType, toolID, out)
	for k, v := range optimized {
		out[k] = v
	}
	return out
}

// CreateAttackChain builds an ordered list of catalog tool names.
func (s *Service) CreateAttackChain(ctx context.Context, target string, objective string) map[string]any {
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	selected := s.SelectTools(ctx, analysis.TargetType, objective)
	steps := make([]map[string]any, 0, len(selected))
	for i, name := range selected {
		steps = append(steps, map[string]any{
			"step": i + 1,
			"tool": name,
		})
	}
	return map[string]any{
		"target":   target,
		"objective": objective,
		"analysis": analysis,
		"steps":    steps,
		"status":   "planned",
	}
}
