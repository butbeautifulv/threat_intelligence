package workeringest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"

	neo4jstore "ds/internal/storage/neo4j"
)

// HandleDSEnvelope applies ds kinds to Neo4j.
func HandleDSEnvelope(ctx context.Context, st *neo4jstore.Store, env *ingestv1.Envelope) error {
	switch env.Kind {
	case ingestv1.KindDSUpsertSigma:
		var p ingestv1.DSUpsertSigmaPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertSigmaRule(ctx, p.ID, p.Title, p.Level, p.LogProduct, p.LogService, p.TagsJSON, p.Markdown, p.Source)
	case ingestv1.KindDSUpsertYara:
		var p ingestv1.DSUpsertYaraPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertYaraRule(ctx, p.ID, p.Name, p.Author, p.TagsJSON, p.Markdown, p.Source)
	case ingestv1.KindDSUpsertAtomic:
		var p ingestv1.DSUpsertAtomicPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertAtomicTest(ctx, p.ID, p.Name, p.Tactic, p.Technique, p.ExecName, p.ExecCmd, p.Markdown, p.Source)
	case ingestv1.KindDSUpsertCaldera:
		var p ingestv1.DSUpsertCalderaPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertCalderaAbility(ctx, p.ID, p.Name, p.Tactic, p.TechniqueID, p.Markdown, p.Source)
	default:
		return fmt.Errorf("ds workeringest: unknown kind %q", env.Kind)
	}
}
