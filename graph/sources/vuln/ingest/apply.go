package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1"

	"github.com/butbeautifulv/threat_intelligence/graph/sources/vuln/internal/domain"
	neo4jstore "github.com/butbeautifulv/threat_intelligence/graph/sources/vuln/storage"
)

// ApplyEnvelope applies vuln kinds to Neo4j.
func ApplyEnvelope(ctx context.Context, st *neo4jstore.Store, env *ingestv1.Envelope) error {
	switch env.Kind {
	case ingestv1.KindVulnUpsert:
		var v domain.Vulnerability
		if err := json.Unmarshal(env.Payload, &v); err != nil {
			return err
		}
		return st.Upsert(ctx, &v)
	case ingestv1.KindVulnMergeExploit:
		var p ingestv1.VulnMergeExploitPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		ref := domain.ExploitRef{Source: p.Source, RefID: p.RefID, URL: p.URL}
		return st.MergeExploitForCVE(ctx, p.CVE, ref)
	default:
		return fmt.Errorf("vuln graph ingest: unknown kind %q", env.Kind)
	}
}
