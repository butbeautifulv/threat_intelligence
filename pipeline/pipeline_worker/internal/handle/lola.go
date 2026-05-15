package handle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"

)

func HandleLola(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case scrapev1.KindLolaArtifactRaw:
		var raw scrapev1.LolaArtifactRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := ingestv1.LolaArtifactPayload{Source: raw.Source, Body: json.RawMessage(raw.RawBody)}
		key := ingestv1.LolaArtifactIdempotencyKey(raw.Source, artifactNameFromBody(raw.RawBody))
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaArtifact, key, pl)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindLolaLoftsRaw:
		var raw scrapev1.LolaLoftsRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := ingestv1.LolaLoftsPayload{Title: raw.Title, Category: raw.Category, LinkURL: raw.LinkURL, Markdown: raw.Markdown}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaLofts, ingestv1.LolaLoftsIdempotencyKey(raw.LinkURL), pl)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindLolaLinkArtifacts:
		pl := ingestv1.LolaLinkArtifactsPayload{}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaLinkArtifacts, ingestv1.LolaLinkArtifactsIdempotencyKey(), pl)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindLolaAttackTechnique:
		var raw scrapev1.LolaAttackTechnique
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := ingestv1.LolaAttackTechniquePayload{ID: raw.ID, Name: raw.Name, Description: raw.Description, Markdown: raw.Markdown}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaAttackTechnique, ingestv1.LolaTechniqueIdempotencyKey(p.ID), p)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindLolaAttackTactic:
		var raw scrapev1.LolaAttackTactic
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := ingestv1.LolaAttackTacticPayload{ID: raw.ID, Name: raw.Name, Description: raw.Description, Markdown: raw.Markdown}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaAttackTactic, ingestv1.LolaTacticIdempotencyKey(p.ID), p)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindLolaMergeTacticTechnique:
		var raw scrapev1.LolaMergeTacticTechnique
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := ingestv1.LolaMergeTacticTechniquePayload{TacticID: raw.TacticID, TechniqueID: raw.TechniqueID}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaMergeTacticTechnique, ingestv1.LolaMergeTacticTechniqueIdempotencyKey(p.TacticID, p.TechniqueID), p)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindLolaMergeSubtechnique:
		var raw scrapev1.LolaMergeSubtechnique
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := ingestv1.LolaMergeSubtechniquePayload{ParentTechniqueID: raw.ParentTechniqueID, ChildTechniqueID: raw.ChildTechniqueID}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceLola, ingestv1.KindLolaMergeSubtechnique, ingestv1.LolaMergeSubtechniqueIdempotencyKey(p.ParentTechniqueID, p.ChildTechniqueID), p)
		return []*ingestv1.Envelope{out}, err

	default:
		return nil, fmt.Errorf("pipeline lola: unknown kind %q", env.Kind)
	}
}

func artifactNameFromBody(body string) string {
	var a struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal([]byte(body), &a)
	return a.Name
}
