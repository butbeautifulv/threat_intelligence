package cve

import "context"

// EnrichCorrelation adds cve_details for parsed indicator CVEs.
func (s *Service) EnrichCorrelation(ctx context.Context, indicators string, related []string) ([]map[string]any, []string) {
	seen := map[string]struct{}{}
	var ids []string
	for _, id := range ParseCVEIDs(indicators) {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	for _, id := range related {
		id = normalizeCVEID(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	var details []map[string]any
	for _, id := range ids {
		details = append(details, s.LookupSummary(ctx, id))
	}
	return details, ids
}
