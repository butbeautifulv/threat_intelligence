package scrapepub

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"scrapepub"

	"github.com/butbeautifulv/threat_intelligence/scrape/sources/lola/internal/domain"
	"github.com/butbeautifulv/threat_intelligence/scrape/sources/lola/internal/repository"
)

type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

type Publisher struct {
	raw rawPublisher
}

var _ repository.LolaRepository = (*Publisher)(nil)

func New(pub *scrapepub.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(scrapepub.NewDomainPublisher(pub, scrapev1.SourceLola, subject))
}

func NewFromRaw(raw rawPublisher) *Publisher {
	return &Publisher{raw: raw}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertArtifact(ctx context.Context, source string, a *domain.Artifact) error {
	if a == nil {
		return nil
	}
	body, err := json.Marshal(a)
	if err != nil {
		return err
	}
	pl := scrapev1.LolaArtifactRaw{Source: source, Path: a.Name, RawBody: string(body)}
	key := "lola:artifact:" + source + ":" + a.Name
	return p.raw.Publish(ctx, scrapev1.KindLolaArtifactRaw, key, pl)
}

func (p *Publisher) UpsertLoftsEntry(ctx context.Context, title, category, linkURL, markdown string) error {
	pl := scrapev1.LolaLoftsRaw{Title: title, Category: category, LinkURL: linkURL, Markdown: markdown}
	return p.raw.Publish(ctx, scrapev1.KindLolaLoftsRaw, "lola:lofts:"+linkURL, pl)
}

func (p *Publisher) UpsertAttackTechnique(ctx context.Context, id, name, description, markdown string) error {
	pl := scrapev1.LolaAttackTechnique{ID: id, Name: name, Description: description, Markdown: markdown}
	return p.raw.Publish(ctx, scrapev1.KindLolaAttackTechnique, "lola:technique:"+id, pl)
}

func (p *Publisher) UpsertAttackTactic(ctx context.Context, id, name, description, markdown string) error {
	pl := scrapev1.LolaAttackTactic{ID: id, Name: name, Description: description, Markdown: markdown}
	return p.raw.Publish(ctx, scrapev1.KindLolaAttackTactic, "lola:tactic:"+id, pl)
}

func (p *Publisher) MergeTacticIncludesTechnique(ctx context.Context, tacticID, techniqueID string) error {
	pl := scrapev1.LolaMergeTacticTechnique{TacticID: tacticID, TechniqueID: techniqueID}
	key := "lola:merge:" + tacticID + ":" + techniqueID
	return p.raw.Publish(ctx, scrapev1.KindLolaMergeTacticTechnique, key, pl)
}

func (p *Publisher) MergeSubtechniqueOf(ctx context.Context, parentTechniqueID, childTechniqueID string) error {
	pl := scrapev1.LolaMergeSubtechnique{ParentTechniqueID: parentTechniqueID, ChildTechniqueID: childTechniqueID}
	key := "lola:sub:" + parentTechniqueID + ":" + childTechniqueID
	return p.raw.Publish(ctx, scrapev1.KindLolaMergeSubtechnique, key, pl)
}

func (p *Publisher) LinkArtifactsAndCommandsToTechniques(ctx context.Context) error {
	return p.raw.Publish(ctx, scrapev1.KindLolaLinkArtifacts, "lola:link-artifacts", struct{}{})
}
