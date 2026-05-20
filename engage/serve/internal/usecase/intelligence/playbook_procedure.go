package intelligence

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
)

const playbookCatalogBoost = 0.12

// attachProcedureContext adds structured playbook catalog tool hints for MITRE ids in blobs.
func attachProcedureContext(ctx context.Context, veil veilgraph.Reader, out map[string]any, blobs ...string) {
	if veil == nil || !veil.Enabled() {
		return
	}
	ids := collectAttackTechniqueIDs(blobs...)
	if len(ids) == 0 {
		return
	}
	allTools := map[string]struct{}{}
	byTechnique := map[string][]string{}
	for _, tid := range ids {
		raw, err := veil.PlaybookRecommendTools(ctx, "", tid)
		if err != nil || !validGraphJSON(raw) {
			continue
		}
		var wrap struct {
			CatalogTools []string `json:"catalog_tools"`
		}
		if json.Unmarshal(raw, &wrap) != nil {
			continue
		}
		byTechnique[tid] = wrap.CatalogTools
		for _, t := range wrap.CatalogTools {
			allTools[t] = struct{}{}
		}
	}
	if len(byTechnique) > 0 {
		out["playbook_catalog_by_technique"] = byTechnique
	}
	if len(allTools) > 0 {
		list := make([]string, 0, len(allTools))
		for t := range allTools {
			list = append(list, t)
		}
		out["playbook_catalog_tools"] = list
	}
}

// playbookCatalogBoost builds decision-engine boost map from veil-api technique tool lists.
func (s *Service) playbookCatalogBoost(ctx context.Context, target, indicators string) map[string]float64 {
	if s.Veil == nil || !s.Veil.Enabled() {
		return nil
	}
	blob := target + " " + indicators
	ids := collectAttackTechniqueIDs(blob)
	if len(ids) == 0 {
		return nil
	}
	boost := map[string]float64{}
	for _, tid := range ids {
		raw, err := s.Veil.PlaybookRecommendTools(ctx, "", tid)
		if err != nil {
			continue
		}
		var wrap struct {
			CatalogTools []string `json:"catalog_tools"`
		}
		if json.Unmarshal(raw, &wrap) != nil {
			continue
		}
		for _, tool := range wrap.CatalogTools {
			boost[tool] = playbookCatalogBoost
		}
	}
	if len(boost) == 0 {
		return nil
	}
	return boost
}
