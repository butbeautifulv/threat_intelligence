// Package natspub publishes ds graph writes as ingestv1 envelopes.
package natspub

import (
	"context"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"
)

type Publisher struct {
	pub     *ingestpub.JetStreamPublisher
	subject string
}

func New(pub *ingestpub.JetStreamPublisher, subject string) *Publisher {
	return &Publisher{pub: pub, subject: strings.TrimSpace(subject)}
}

func (p *Publisher) publish(ctx context.Context, kind, idem string, payload any) error {
	env, err := ingestv1.NewEnvelope(ingestv1.SourceDS, kind, idem, payload)
	if err != nil {
		return err
	}
	return p.pub.PublishJSON(ctx, p.subject, env)
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertSigmaRule(ctx context.Context, id, title, level, logProduct, logService, tagsJSON, markdown, source string) error {
	pl := ingestv1.DSUpsertSigmaPayload{
		ID: id, Title: title, Level: level, LogProduct: logProduct, LogService: logService,
		TagsJSON: tagsJSON, Markdown: markdown, Source: source,
	}
	return p.publish(ctx, ingestv1.KindDSUpsertSigma, ingestv1.DSSigmaIdempotencyKey(id), pl)
}

func (p *Publisher) UpsertYaraRule(ctx context.Context, id, name, author, tagsJSON, markdown, source string) error {
	pl := ingestv1.DSUpsertYaraPayload{ID: id, Name: name, Author: author, TagsJSON: tagsJSON, Markdown: markdown, Source: source}
	return p.publish(ctx, ingestv1.KindDSUpsertYara, ingestv1.DSYaraIdempotencyKey(id), pl)
}

func (p *Publisher) UpsertAtomicTest(ctx context.Context, id, name, tactic, technique, execName, execCmd, markdown, source string) error {
	pl := ingestv1.DSUpsertAtomicPayload{
		ID: id, Name: name, Tactic: tactic, Technique: technique, ExecName: execName, ExecCmd: execCmd, Markdown: markdown, Source: source,
	}
	return p.publish(ctx, ingestv1.KindDSUpsertAtomic, ingestv1.DSAtomicIdempotencyKey(id), pl)
}

func (p *Publisher) UpsertCalderaAbility(ctx context.Context, id, name, tactic, techniqueID, markdown, source string) error {
	pl := ingestv1.DSUpsertCalderaPayload{
		ID: id, Name: name, Tactic: tactic, TechniqueID: techniqueID, Markdown: markdown, Source: source,
	}
	return p.publish(ctx, ingestv1.KindDSUpsertCaldera, ingestv1.DSCalderaIdempotencyKey(id), pl)
}
