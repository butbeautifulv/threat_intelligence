package intelligence

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/decision"
)

func (s *Service) candidateIDs(targetType string) []string {
	return s.engine().CandidateTools(targetType)
}

func filterEnabled(names []string, reg *tools.Registry) []string {
	if reg == nil {
		return names
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		spec, ok := reg.Get(name)
		if ok && spec.Enabled {
			out = append(out, name)
		}
	}
	return out
}

func catalogToShortID(catalogName string) string {
	for short, full := range tools.BinaryToCatalog {
		if full == catalogName {
			return short
		}
	}
	return catalogName
}

func (s *Service) SelectTools(ctx context.Context, targetType, objective string) []string {
	return s.SelectToolsForTarget(ctx, targetType, objective, "")
}

func (s *Service) SelectToolsForTarget(ctx context.Context, targetType, objective, target string) []string {
	cands := s.candidateIDs(targetType)
	_, techLabels, cms, _, _, _ := probeTarget(ctx, target)
	stack := labelsToTechnologies(techLabels, cms)
	boost := mergeBoost(s.graphBoost(ctx, target), techStackBoost(stack), cmsToolBoost(cms, s.Registry))
	ranked := s.engine().RankToolsWithBoost(targetType, cands, boost)
	ranked = appendTechSpecificTools(ranked, stack, cms)
	ranked = s.engine().RankToolsWithBoost(targetType, ranked, boost)
	names := tools.ResolveCatalogNames(ranked, s.Registry)
	names = filterEnabled(names, s.Registry)
	obj := strings.ToLower(strings.TrimSpace(objective))
	if obj == "stealth" {
		ids := ranked
		names = tools.ResolveCatalogNames(decision.FilterStealthTools(ids), s.Registry)
		names = filterEnabled(names, s.Registry)
		return decision.CapTools(names, objective)
	}
	if obj == "comprehensive" {
		filtered := decision.FilterComprehensiveTools(s.engine(), targetType, ranked)
		if len(filtered) > 0 {
			names = tools.ResolveCatalogNames(filtered, s.Registry)
			names = filterEnabled(names, s.Registry)
		}
	}
	if obj == "quick" || obj == "fast" {
		if len(names) > 3 {
			names = names[:3]
		}
		return names
	}
	return decision.CapToolsWithEngine(names, targetType, objective, s.engine(), catalogToShortID)
}

func appendTechSpecificTools(ranked []string, stack []Technology, cms string) []string {
	seen := map[string]struct{}{}
	for _, id := range ranked {
		seen[id] = struct{}{}
	}
	add := func(id string) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ranked = append(ranked, id)
	}
	for _, t := range stack {
		switch t {
		case TechWordPress:
			add("wpscan")
		case TechPHP:
			add("nikto")
		}
	}
	if strings.EqualFold(cms, "wordpress") {
		add("wpscan")
	}
	return ranked
}

func labelsToTechnologies(labels []string, cms string) []Technology {
	if cms != "" {
		labels = append(labels, cms)
	}
	seen := map[Technology]struct{}{}
	for _, l := range labels {
		switch strings.ToLower(l) {
		case "wordpress":
			seen[TechWordPress] = struct{}{}
		case "drupal":
			seen[TechDrupal] = struct{}{}
		case "joomla":
			seen[TechJoomla] = struct{}{}
		case "php":
			seen[TechPHP] = struct{}{}
		case "nginx":
			seen[TechNginx] = struct{}{}
		case "apache":
			seen[TechApache] = struct{}{}
		case "nodejs":
			seen[TechNodeJS] = struct{}{}
		case "java":
			seen[TechJava] = struct{}{}
		case "dotnet":
			seen[TechDotNet] = struct{}{}
		}
	}
	out := make([]Technology, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	return out
}

func cmsToolBoost(cms string, reg *tools.Registry) map[string]float64 {
	if cms == "" || reg == nil {
		return nil
	}
	boost := map[string]float64{}
	switch cms {
	case "wordpress":
		boost["wpscan"] = 0.25
		boost["nuclei"] = 0.05
	case "php":
		boost["nikto"] = 0.15
		boost["sqlmap"] = 0.12
	case "drupal", "joomla":
		boost["nuclei"] = 0.1
		boost["nikto"] = 0.1
	}
	for id := range boost {
		name := tools.ResolveCatalogName(id, reg)
		spec, ok := reg.Get(name)
		if !ok || !spec.Enabled {
			delete(boost, id)
		}
	}
	return boost
}

func mergeBoost(parts ...map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for _, p := range parts {
		for k, v := range p {
			out[k] += v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
