// Package natspub implements repository.LolaRepository via NATS publish.
package natspub

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"

	"lola/internal/domain"
	"lola/internal/repository"
)

type Publisher struct {
	pub     *ingestpub.JetStreamPublisher
	subject string
}

var _ repository.LolaRepository = (*Publisher)(nil)

func New(pub *ingestpub.JetStreamPublisher, subject string) *Publisher {
	return &Publisher{pub: pub, subject: strings.TrimSpace(subject)}
}

func (p *Publisher) publish(ctx context.Context, kind, idem string, payload any) error {
	env, err := ingestv1.NewEnvelope(ingestv1.SourceLola, kind, idem, payload)
	if err != nil {
		return err
	}
	return p.pub.PublishJSON(ctx, p.subject, env)
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertArtifact(ctx context.Context, source string, a *domain.Artifact) error {
	if a == nil {
		return nil
	}
	raw, err := json.Marshal(a)
	if err != nil {
		return err
	}
	pl := ingestv1.LolaArtifactPayload{Source: source, Body: raw}
	key := ingestv1.LolaArtifactIdempotencyKey(source, a.Name)
	return p.publish(ctx, ingestv1.KindLolaArtifact, key, pl)
}

func (p *Publisher) UpsertLoftsEntry(ctx context.Context, title, category, linkURL, markdown string) error {
	pl := ingestv1.LolaLoftsPayload{Title: title, Category: category, LinkURL: linkURL, Markdown: markdown}
	return p.publish(ctx, ingestv1.KindLolaLofts, ingestv1.LolaLoftsIdempotencyKey(linkURL), pl)
}

func (p *Publisher) UpsertAttackTechnique(ctx context.Context, id, name, description, markdown string) error {
	pl := ingestv1.LolaAttackTechniquePayload{ID: id, Name: name, Description: description, Markdown: markdown}
	return p.publish(ctx, ingestv1.KindLolaAttackTechnique, ingestv1.LolaTechniqueIdempotencyKey(id), pl)
}

func (p *Publisher) UpsertAttackTactic(ctx context.Context, id, name, description, markdown string) error {
	pl := ingestv1.LolaAttackTacticPayload{ID: id, Name: name, Description: description, Markdown: markdown}
	return p.publish(ctx, ingestv1.KindLolaAttackTactic, ingestv1.LolaTacticIdempotencyKey(id), pl)
}

func (p *Publisher) MergeTacticIncludesTechnique(ctx context.Context, tacticID, techniqueID string) error {
	pl := ingestv1.LolaMergeTacticTechniquePayload{TacticID: tacticID, TechniqueID: techniqueID}
	return p.publish(ctx, ingestv1.KindLolaMergeTacticTechnique, ingestv1.LolaMergeTacticTechniqueIdempotencyKey(tacticID, techniqueID), pl)
}

func (p *Publisher) MergeSubtechniqueOf(ctx context.Context, parentTechniqueID, childTechniqueID string) error {
	pl := ingestv1.LolaMergeSubtechniquePayload{ParentTechniqueID: parentTechniqueID, ChildTechniqueID: childTechniqueID}
	return p.publish(ctx, ingestv1.KindLolaMergeSubtechnique, ingestv1.LolaMergeSubtechniqueIdempotencyKey(parentTechniqueID, childTechniqueID), pl)
}

func (p *Publisher) LinkArtifactsAndCommandsToTechniques(ctx context.Context) error {
	pl := ingestv1.LolaLinkArtifactsPayload{}
	return p.publish(ctx, ingestv1.KindLolaLinkArtifacts, ingestv1.LolaLinkArtifactsIdempotencyKey(), pl)
}
