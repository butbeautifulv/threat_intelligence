// Package scrapepub publishes ds raw fetches to scrape.> (pipeline-worker → ingest.>).
package scrapepub

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"scrapepub"
)

type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

// Publisher implements graphStore via a raw scrape publisher.
type Publisher struct {
	raw rawPublisher
}

func New(pub *scrapepub.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(scrapepub.NewDomainPublisher(pub, scrapev1.SourceDS, subject))
}

// NewFromRaw wraps any publisher that emits scrapev1 (e.g. factory RawPublisher).
func NewFromRaw(raw rawPublisher) *Publisher {
	return &Publisher{raw: raw}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertSigmaRaw(ctx context.Context, path, rawYAML string) error {
	pl := scrapev1.DSSigmaRaw{Path: path, RawYAML: rawYAML}
	return p.raw.Publish(ctx, scrapev1.KindDSSigmaRaw, scrapev1.DSContentKey("sigma", path), pl)
}

func (p *Publisher) UpsertYaraRaw(ctx context.Context, path, name, rawBody string) error {
	pl := scrapev1.DSYaraRaw{Path: path, Name: name, RawBody: rawBody}
	return p.raw.Publish(ctx, scrapev1.KindDSYaraRaw, scrapev1.DSContentKey("yara", path), pl)
}

func (p *Publisher) UpsertAtomicRaw(ctx context.Context, path, rawYAML string) error {
	pl := scrapev1.DSAtomicRaw{Path: path, RawYAML: rawYAML}
	return p.raw.Publish(ctx, scrapev1.KindDSAtomicRaw, scrapev1.DSContentKey("atomic", path), pl)
}

func (p *Publisher) UpsertCalderaRaw(ctx context.Context, path, fileName, rawBody string) error {
	pl := scrapev1.DSCalderaRaw{Path: path, FileName: fileName, RawBody: rawBody}
	return p.raw.Publish(ctx, scrapev1.KindDSCalderaRaw, scrapev1.DSContentKey("caldera", path), pl)
}
