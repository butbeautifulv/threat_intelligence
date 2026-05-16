package cve

import (
	"context"
	"strings"
	"time"
)

// MonitorFeeds fetches recent CVEs and analyzes top entries.
func (s *Service) MonitorFeeds(ctx context.Context, req MonitorRequest) MonitorResult {
	now := time.Now().UTC().Format(time.RFC3339)
	if req.Hours <= 0 {
		req.Hours = 24
	}
	if req.AnalyzeTop <= 0 {
		req.AnalyzeTop = 5
	}
	if s.NVD == nil {
		return MonitorResult{Success: false, Error: "NVD client not configured", Timestamp: now}
	}
	entries, err := s.NVD.FetchRecent(ctx, req.Hours, req.SeverityFilter)
	if err != nil {
		return MonitorResult{
			Success:   false,
			Error:     err.Error(),
			Timestamp: now,
			CVEMonitoring: map[string]any{
				"success": false,
				"cves":    []CVEEntry{},
			},
		}
	}
	if kw := strings.TrimSpace(req.Keywords); kw != "" {
		entries = filterByKeywords(entries, kw)
	}
	cves := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		cves = append(cves, entryToMap(e))
	}
	monitoring := map[string]any{
		"success":       true,
		"cves":          cves,
		"total_results": len(cves),
		"hours":         req.Hours,
		"severity_filter": req.SeverityFilter,
	}
	if req.Keywords != "" {
		monitoring["filtered_by_keywords"] = req.Keywords
		monitoring["total_after_filter"] = len(cves)
	}
	var analyses []ExploitabilityAnalysis
	top := req.AnalyzeTop
	if top > len(entries) {
		top = len(entries)
	}
	for i := 0; i < top; i++ {
		analyses = append(analyses, AnalyzeExploitability(entries[i]))
	}
	var veilEnrich []VeilEnrichment
	if s.Veil != nil && s.Veil.Enabled() {
		ids := make([]string, 0, top)
		for i := 0; i < top; i++ {
			ids = append(ids, entries[i].CVEID)
		}
		veilEnrich = EnrichFromVeil(ctx, s.Veil, ids)
	}
	out := MonitorResult{
		Success:                true,
		CVEMonitoring:          monitoring,
		ExploitabilityAnalysis: analyses,
		Timestamp:              now,
	}
	if len(veilEnrich) > 0 {
		monitoring["veil_enrichment"] = veilEnrich
	}
	return out
}

func filterByKeywords(entries []CVEEntry, keywords string) []CVEEntry {
	parts := strings.Split(strings.ToLower(keywords), ",")
	var out []CVEEntry
	for _, e := range entries {
		desc := strings.ToLower(e.Description)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" && strings.Contains(desc, p) {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

func entryToMap(e CVEEntry) map[string]any {
	return map[string]any{
		"cve_id":      e.CVEID,
		"description": e.Description,
		"severity":    e.Severity,
		"cvss_score":  e.CVSSScore,
		"references":  e.References,
	}
}
