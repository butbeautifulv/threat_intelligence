package cve

import (
	"context"
	"time"
)

// Lookup fetches a single CVE with exploitability analysis and optional veil enrichment.
func (s *Service) Lookup(ctx context.Context, cveID string) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	cveID = normalizeCVEID(cveID)
	if cveID == "" {
		return map[string]any{"success": false, "error": "invalid cve_id", "timestamp": now}
	}
	if s.NVD == nil {
		return map[string]any{"success": false, "error": "NVD client not configured", "timestamp": now}
	}
	entry, err := s.NVD.FetchCVE(ctx, cveID)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error(), "cve_id": cveID, "timestamp": now}
	}
	analysis := AnalyzeExploitability(*entry)
	out := map[string]any{
		"success":     true,
		"cve":         entryToMap(*entry),
		"analysis":    analysis,
		"timestamp":   now,
	}
	if s.Veil != nil && s.Veil.Enabled() {
		enrich := EnrichFromVeil(ctx, s.Veil, []string{cveID})
		if len(enrich) > 0 {
			out["veil_enrichment"] = enrich
		}
	}
	return out
}
