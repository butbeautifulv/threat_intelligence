package scrapepub

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/harvest"
	connats "github.com/butbeautifulv/veil/scrape/connector/nats"

	"github.com/butbeautifulv/veil/pkg/vuln/domain"
	"github.com/butbeautifulv/veil/scrape/harvest/internal/sources/vuln/internal/repository"
)

type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

type Publisher struct {
	raw rawPublisher
}

var _ repository.VulnerabilityRepository = (*Publisher)(nil)

func New(pub *connats.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(connats.NewDomainPublisher(pub, harvest.SourceVuln, subject))
}

func NewFromRaw(raw rawPublisher) *Publisher {
	return &Publisher{raw: raw}
}

func (p *Publisher) Save(ctx context.Context, v *domain.Vulnerability) error {
	return p.Upsert(ctx, v)
}

func (p *Publisher) FindByCVE(_ context.Context, _ string) (*domain.Vulnerability, error) {
	return nil, errors.New("vuln scrapepub: FindByCVE not supported")
}

func (p *Publisher) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	if v == nil || strings.TrimSpace(v.CVE) == "" {
		return fmt.Errorf("vuln scrapepub: empty CVE")
	}
	key := "vuln:cve:" + strings.TrimSpace(strings.ToUpper(v.CVE))
	return p.raw.Publish(ctx, harvest.KindVulnCVEUpsert, key, v)
}

func (p *Publisher) MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error {
	cve = strings.TrimSpace(strings.ToUpper(cve))
	if cve == "" {
		return nil
	}
	pl := harvest.VulnMergeExploit{CVE: cve, Source: ref.Source, RefID: ref.RefID, URL: ref.URL}
	key := "vuln:exploit:" + cve + ":" + ref.Source + ":" + ref.RefID
	return p.raw.Publish(ctx, harvest.KindVulnMergeExploit, key, pl)
}

func (p *Publisher) PublishNVDPage(ctx context.Context, startIndex int, rawJSON []byte) error {
	pl := harvest.VulnNVDPage{StartIndex: startIndex, RawJSON: string(rawJSON)}
	key := fmt.Sprintf("vuln:nvd:page:%d", startIndex)
	return p.raw.Publish(ctx, harvest.KindVulnNVDPage, key, pl)
}
