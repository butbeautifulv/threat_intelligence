package cve

import (
	"context"
	"encoding/json"
)

// VeilSearcher is implemented by veilgraph.Client.
type VeilSearcher interface {
	Enabled() bool
	Search(ctx context.Context, category, query string) (json.RawMessage, error)
}

// VeilEnrichment is graph context for a CVE id.
type VeilEnrichment struct {
	CVEID  string          `json:"cve_id"`
	Raw    json.RawMessage `json:"raw,omitempty"`
	Found  bool            `json:"found"`
}

// EnrichFromVeil queries veil vuln category for each CVE id.
func EnrichFromVeil(ctx context.Context, veil VeilSearcher, cveIDs []string) []VeilEnrichment {
	if veil == nil || !veil.Enabled() {
		return nil
	}
	var out []VeilEnrichment
	for _, id := range cveIDs {
		if id == "" {
			continue
		}
		e := VeilEnrichment{CVEID: id}
		raw, err := veil.Search(ctx, "vuln", id)
		if err == nil && len(raw) > 2 && string(raw) != "null" {
			e.Raw = raw
			e.Found = true
		}
		out = append(out, e)
	}
	return out
}
