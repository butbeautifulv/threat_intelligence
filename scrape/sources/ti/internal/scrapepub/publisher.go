// Package scrapepub publishes raw TI events to scrape.> for pipeline-worker.
package scrapepub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
	"scrapepub"

	"github.com/butbeautifulv/threat_intelligence/scrape/sources/ti/internal/domain"
	"github.com/butbeautifulv/threat_intelligence/scrape/sources/ti/internal/repository"
)

type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

type Publisher struct {
	raw rawPublisher
}

var _ repository.GraphRepository = (*Publisher)(nil)

func New(pub *scrapepub.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(scrapepub.NewDomainPublisher(pub, scrapev1.SourceTI, subject))
}

func NewFromRaw(raw rawPublisher) *Publisher {
	return &Publisher{raw: raw}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertIOC(ctx context.Context, i domain.IOC) error {
	key := "ti:ioc:" + strings.TrimSpace(string(i.Type)) + ":" + strings.TrimSpace(i.Value)
	return p.raw.Publish(ctx, scrapev1.KindTIIoCRaw, key, i)
}

func (p *Publisher) UpsertCampaign(ctx context.Context, c domain.Campaign) error {
	key := "ti:campaign:" + strings.TrimSpace(c.ID)
	return p.raw.Publish(ctx, scrapev1.KindTICampaignRaw, key, c)
}

func (p *Publisher) UpsertCluster(ctx context.Context, cl domain.Cluster) error {
	key := "ti:cluster:" + strings.TrimSpace(cl.ID)
	return p.raw.Publish(ctx, scrapev1.KindTIClusterRaw, key, cl)
}

func (p *Publisher) UpsertActor(ctx context.Context, a domain.Actor) error {
	key := "ti:actor:" + strings.TrimSpace(a.Name)
	return p.raw.Publish(ctx, scrapev1.KindTIActorRaw, key, a)
}

func (p *Publisher) UpsertReport(ctx context.Context, r domain.Report) error {
	key := "ti:report:" + strings.TrimSpace(r.Link)
	return p.raw.Publish(ctx, scrapev1.KindTIReportRaw, key, r)
}

func (p *Publisher) LinkCampaignIOC(ctx context.Context, campaignID string, i domain.IOC) error {
	_ = campaignID
	return p.UpsertIOC(ctx, i)
}

func (p *Publisher) LinkClusterCampaign(ctx context.Context, clusterID, campaignID string) error {
	_ = clusterID
	_ = campaignID
	return nil
}

func (p *Publisher) LinkCampaignActor(ctx context.Context, campaignID, actorName string) error {
	_ = campaignID
	return p.UpsertActor(ctx, domain.Actor{Name: actorName})
}

func (p *Publisher) LinkReportMentionsIOC(ctx context.Context, reportID string, i domain.IOC) error {
	_ = reportID
	return p.UpsertIOC(ctx, i)
}

func (p *Publisher) UpsertKEVVulnerability(ctx context.Context, cve, vendor, product, summary, dateAdded string) error {
	pl := scrapev1.TIKEVRow{
		CVEID: cve, VendorProject: vendor, Product: product, ShortDesc: summary, DateAdded: dateAdded,
	}
	cveU := strings.TrimSpace(strings.ToUpper(cve))
	return p.raw.Publish(ctx, scrapev1.KindTIKEVRow, "ti:kev:"+cveU, pl)
}

// PublishJSONLLine publishes one raw JSONL line.
func (p *Publisher) PublishJSONLLine(ctx context.Context, line []byte) error {
	sum := sha256.Sum256(line)
	key := "ti:jsonl:" + hex.EncodeToString(sum[:])
	pl := scrapev1.TIJSONLLine{Line: json.RawMessage(line)}
	return p.raw.Publish(ctx, scrapev1.KindTIJSONLLine, key, pl)
}
