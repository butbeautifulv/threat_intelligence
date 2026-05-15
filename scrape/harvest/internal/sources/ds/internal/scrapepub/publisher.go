// Package scrapepub publishes ds raw fetches to scrape.> (pipeline-worker → ingest.>).
package scrapepub

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/pkg/harvest"
	connats "github.com/butbeautifulv/threat_intelligence/scrape/connector/nats"
)

type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

// Publisher implements graphStore via a raw scrape publisher.
type Publisher struct {
	raw rawPublisher
}

func New(pub *connats.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(connats.NewDomainPublisher(pub, harvest.SourceDS, subject))
}

// NewFromRaw wraps any publisher that emits harvest (e.g. factory RawPublisher).
func NewFromRaw(raw rawPublisher) *Publisher {
	return &Publisher{raw: raw}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertSigmaRaw(ctx context.Context, path, rawYAML string) error {
	pl := harvest.DSSigmaRaw{Path: path, RawYAML: rawYAML}
	return p.raw.Publish(ctx, harvest.KindDSSigmaRaw, harvest.DSContentKey("sigma", path), pl)
}

func (p *Publisher) UpsertYaraRaw(ctx context.Context, path, name, rawBody string) error {
	pl := harvest.DSYaraRaw{Path: path, Name: name, RawBody: rawBody}
	return p.raw.Publish(ctx, harvest.KindDSYaraRaw, harvest.DSContentKey("yara", path), pl)
}

func (p *Publisher) UpsertAtomicRaw(ctx context.Context, path, rawYAML string) error {
	pl := harvest.DSAtomicRaw{Path: path, RawYAML: rawYAML}
	return p.raw.Publish(ctx, harvest.KindDSAtomicRaw, harvest.DSContentKey("atomic", path), pl)
}

func (p *Publisher) UpsertCalderaRaw(ctx context.Context, path, fileName, rawBody string) error {
	pl := harvest.DSCalderaRaw{Path: path, FileName: fileName, RawBody: rawBody}
	return p.raw.Publish(ctx, harvest.KindDSCalderaRaw, harvest.DSContentKey("caldera", path), pl)
}
