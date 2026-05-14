// Package natspub implements vuln VulnerabilityRepository by publishing ingestv1 envelopes.
package natspub

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"

	"vuln/internal/domain"
	"vuln/internal/repository"
)

type Publisher struct {
	pub     *ingestpub.JetStreamPublisher
	subject string
}

var _ repository.VulnerabilityRepository = (*Publisher)(nil)

func New(pub *ingestpub.JetStreamPublisher, subject string) *Publisher {
	return &Publisher{pub: pub, subject: strings.TrimSpace(subject)}
}

func (p *Publisher) Save(ctx context.Context, v *domain.Vulnerability) error {
	return p.Upsert(ctx, v)
}

func (p *Publisher) FindByCVE(_ context.Context, _ string) (*domain.Vulnerability, error) {
	return nil, errors.New("vuln natspub: FindByCVE not supported in producer mode")
}

func (p *Publisher) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	if v == nil || strings.TrimSpace(v.CVE) == "" {
		return fmt.Errorf("vuln natspub: empty vulnerability or CVE")
	}
	key := ingestv1.VulnUpsertIdempotencyKey(v.CVE)
	env, err := ingestv1.NewEnvelope(ingestv1.SourceVuln, ingestv1.KindVulnUpsert, key, v)
	if err != nil {
		return err
	}
	return p.pub.PublishJSON(ctx, p.subject, env)
}

func (p *Publisher) MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error {
	cve = strings.TrimSpace(strings.ToUpper(cve))
	if cve == "" {
		return nil
	}
	pl := ingestv1.VulnMergeExploitPayload{CVE: cve, Source: ref.Source, RefID: ref.RefID, URL: ref.URL}
	key := ingestv1.VulnMergeExploitIdempotencyKey(cve, ref.Source, ref.RefID)
	env, err := ingestv1.NewEnvelope(ingestv1.SourceVuln, ingestv1.KindVulnMergeExploit, key, pl)
	if err != nil {
		return err
	}
	return p.pub.PublishJSON(ctx, p.subject, env)
}
