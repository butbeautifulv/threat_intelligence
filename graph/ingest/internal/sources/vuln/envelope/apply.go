package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"

	"github.com/butbeautifulv/veil/graph/ingest/internal/sources/vuln/domain"
	neo4jstore "github.com/butbeautifulv/veil/graph/ingest/internal/sources/vuln/storage"
)

// ApplyEnvelope applies vuln kinds to Neo4j.
func ApplyEnvelope(ctx context.Context, st *neo4jstore.Store, env *commit.Envelope) error {
	switch env.Kind {
	case commit.KindVulnUpsert:
		var v domain.Vulnerability
		if err := json.Unmarshal(env.Payload, &v); err != nil {
			return err
		}
		return st.Upsert(ctx, &v)
	case commit.KindVulnMergeExploit:
		var p commit.VulnMergeExploitPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		ref := domain.ExploitRef{Source: p.Source, RefID: p.RefID, URL: p.URL}
		return st.MergeExploitForCVE(ctx, p.CVE, ref)
	default:
		return fmt.Errorf("vuln graph ingest: unknown kind %q", env.Kind)
	}
}
