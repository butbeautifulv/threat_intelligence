package intelligence

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ParallelToolRunner runs multiple catalog tools concurrently (workflow.Service).
type ParallelToolRunner interface {
	RunToolsParallel(ctx context.Context, subject, target, targetType string, toolNames []string) []map[string]any
}

// Service provides target analysis and tool selection.
type Service struct {
	Veil           veilgraph.Reader
	Registry       *tools.Registry
	Engine         *DecisionEngine
	Tools          *toolsuc.Runner
	Audit          audit.Reader
	CVE            CVEIntelligence
	ParallelRunner ParallelToolRunner
}

// CVEIntelligence is implemented by cve.Service (defined here to avoid import cycles in tests).
type CVEIntelligence interface {
	EnrichCorrelation(ctx context.Context, indicators string, related []string) ([]map[string]any, []string)
	BuildCVEAttackPaths(ctx context.Context, cveIDs []string) []map[string]any
}

func (s *Service) engine() *DecisionEngine {
	if s.Engine != nil {
		return s.Engine
	}
	return DefaultDecisionEngine()
}

func (s *Service) AnalyzeTarget(ctx context.Context, req contract.AnalyzeTargetRequest) contract.AnalyzeTargetResponse {
	target := strings.TrimSpace(req.Target)
	tt, tech, cms, _, probeHdr, probeBody := probeTarget(ctx, target)
	ips := resolveTargetIPs(ctx, target)
	techLabels := technologiesDetected(tech, cms)
	profile := BuildTargetProfile(target, tt, techLabels, cms, ips, 0)
	resp := contract.AnalyzeTargetResponse{
		Target:       target,
		TargetType:   tt,
		Technologies: tech,
		RiskLevel:    profile.RiskLevel,
		Confidence:   profile.ConfidenceScore,
		Metadata:     map[string]any{},
	}
	if cms != "" {
		resp.Metadata["cms"] = cms
	}
	resp.Metadata["attack_surface_score"] = profile.AttackSurfaceScore
	resp.Metadata["ip_addresses"] = profile.IPAddresses
	stack := DetectTechnologies(ctx, target, probeHdr, probeBody)
	resp.Metadata["technologies_detected"] = TechnologiesToStrings(stack)
	resp.Metadata["technology_stack"] = TechnologiesToStrings(stack)
	if s.Veil != nil && s.Veil.Enabled() {
		if raw, err := s.Veil.Categories(ctx); err == nil {
			resp.Metadata["veil_categories"] = raw
			if resp.Confidence < 0.85 {
				resp.Confidence = 0.85
			}
		}
		s.enrichGraph(ctx, target, resp.Metadata)
		if boost := s.graphBoost(ctx, target); len(boost) > 0 {
			resp.Metadata["graph_tool_boost"] = boost
		}
	}
	return resp
}

// BuildProfileFromTarget runs probe + profile assembly for internal engine use.
func (s *Service) BuildProfileFromTarget(ctx context.Context, target string) TargetProfile {
	tt, tech, cms, _, _, _ := probeTarget(ctx, target)
	ips := resolveTargetIPs(ctx, target)
	labels := technologiesDetected(tech, cms)
	return BuildTargetProfile(target, tt, labels, cms, ips, 0)
}

func (s *Service) enrichGraph(ctx context.Context, target string, meta map[string]any) {
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{})
	if len(state.Hits) > 0 {
		meta["graph_hits"] = state.Hits
		meta["graph_vuln_context"] = true
		meta["graph_host"] = state.Host
	}
}

func (s *Service) graphBoost(ctx context.Context, target string) map[string]float64 {
	state := LoadTargetGraph(ctx, s.Veil, target, TargetGraphLoadOpts{})
	if !state.GraphEnabled || state.Host == "" {
		return nil
	}
	boost := map[string]float64{}
	for cat, tools := range map[string][]string{
		"vuln":   {"nuclei", "nikto", "sqlmap"},
		"ti":     {"nuclei", "httpx"},
		"engage": {"nuclei", "nmap", "httpx"},
	} {
		if _, ok := state.Hits[cat]; !ok {
			continue
		}
		for _, t := range tools {
			boost[t] += 0.08
		}
	}
	if len(boost) == 0 {
		return nil
	}
	return boost
}

// TechnologyDetection returns technologies, CMS, and confidence for a target.
func (s *Service) TechnologyDetection(ctx context.Context, target string) map[string]any {
	analysis := s.AnalyzeTarget(ctx, contract.AnalyzeTargetRequest{Target: target})
	cms, _ := analysis.Metadata["cms"].(string)
	stackRaw, _ := analysis.Metadata["technology_stack"].([]string)
	if stackRaw == nil {
		if ts, ok := analysis.Metadata["technology_stack"].([]any); ok {
			for _, v := range ts {
				if str, ok := v.(string); ok {
					stackRaw = append(stackRaw, str)
				}
			}
		}
	}
	return map[string]any{
		"target":           analysis.Target,
		"target_type":      analysis.TargetType,
		"technologies":     analysis.Technologies,
		"technology_stack": stackRaw,
		"cms":              cms,
		"confidence":       analysis.Confidence,
		"risk_level":       analysis.RiskLevel,
	}
}
