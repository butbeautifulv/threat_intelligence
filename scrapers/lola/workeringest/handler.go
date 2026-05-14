package workeringest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"

	"lola/internal/domain"
	neo4jstore "lola/internal/storage/neo4j"
)

// HandleLolaEnvelope applies lola kinds to Neo4j.
func HandleLolaEnvelope(ctx context.Context, st *neo4jstore.Store, env *ingestv1.Envelope) error {
	switch env.Kind {
	case ingestv1.KindLolaArtifact:
		var p ingestv1.LolaArtifactPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		var a domain.Artifact
		if err := json.Unmarshal(p.Body, &a); err != nil {
			return err
		}
		return st.UpsertArtifact(ctx, p.Source, &a)
	case ingestv1.KindLolaLofts:
		var p ingestv1.LolaLoftsPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertLoftsEntry(ctx, p.Title, p.Category, p.LinkURL, p.Markdown)
	case ingestv1.KindLolaAttackTechnique:
		var p ingestv1.LolaAttackTechniquePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertAttackTechnique(ctx, p.ID, p.Name, p.Description, p.Markdown)
	case ingestv1.KindLolaAttackTactic:
		var p ingestv1.LolaAttackTacticPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertAttackTactic(ctx, p.ID, p.Name, p.Description, p.Markdown)
	case ingestv1.KindLolaMergeTacticTechnique:
		var p ingestv1.LolaMergeTacticTechniquePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.MergeTacticIncludesTechnique(ctx, p.TacticID, p.TechniqueID)
	case ingestv1.KindLolaMergeSubtechnique:
		var p ingestv1.LolaMergeSubtechniquePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.MergeSubtechniqueOf(ctx, p.ParentTechniqueID, p.ChildTechniqueID)
	case ingestv1.KindLolaLinkArtifacts:
		return st.LinkArtifactsAndCommandsToTechniques(ctx)
	default:
		return fmt.Errorf("lola workeringest: unknown kind %q", env.Kind)
	}
}
