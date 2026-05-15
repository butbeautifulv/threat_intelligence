// Package natspub implements repository.GraphRepository by publishing ingestv1 envelopes to NATS.
package natspub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"
	"ingestpub"

	"ti/internal/domain"
	"ti/internal/normalize"
	"ti/internal/repository"
)

// Publisher publishes TI graph operations as JetStream messages.
type Publisher struct {
	pub     *ingestpub.JetStreamPublisher
	subject string
}

var _ repository.GraphRepository = (*Publisher)(nil)

func New(pub *ingestpub.JetStreamPublisher, subject string) *Publisher {
	return &Publisher{pub: pub, subject: strings.TrimSpace(subject)}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) publish(ctx context.Context, kind, idem string, payload any) error {
	env, err := ingestv1.NewEnvelope(ingestv1.SourceTI, kind, idem, payload)
	if err != nil {
		return err
	}
	return p.pub.PublishJSON(ctx, p.subject, env)
}

func (p *Publisher) UpsertIOC(ctx context.Context, i domain.IOC) error {
	ni, ok := normalize.NormalizeIOC(i)
	if !ok {
		return nil
	}
	id := normalize.CanonicalID(ni)
	return p.publish(ctx, ingestv1.KindTIIoC, ingestv1.TIIoCIdempotencyKey(id), ni)
}

func (p *Publisher) UpsertCampaign(ctx context.Context, c domain.Campaign) error {
	c = normalize.NormalizeCampaign(c)
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("campaign requires id and name")
	}
	return p.publish(ctx, ingestv1.KindTICampaign, ingestv1.TICampaignIdempotencyKey(c.ID), c)
}

func (p *Publisher) UpsertCluster(ctx context.Context, cl domain.Cluster) error {
	cl = normalize.NormalizeCluster(cl)
	if cl.ID == "" || cl.Name == "" {
		return fmt.Errorf("cluster requires id and name")
	}
	return p.publish(ctx, ingestv1.KindTICluster, ingestv1.TIClusterIdempotencyKey(cl.ID), cl)
}

func (p *Publisher) UpsertActor(ctx context.Context, a domain.Actor) error {
	id := strings.TrimSpace(a.ID)
	if id == "" {
		id = normalize.ActorStableID(a.Name)
	}
	return p.publish(ctx, ingestv1.KindTIActor, ingestv1.TIActorIdempotencyKey(id), a)
}

func (p *Publisher) UpsertReport(ctx context.Context, r domain.Report) error {
	sid := normalize.ReportStableID(r.Link)
	return p.publish(ctx, ingestv1.KindTIReport, ingestv1.TIReportIdempotencyKey(sid), r)
}

func (p *Publisher) LinkCampaignIOC(ctx context.Context, campaignID string, i domain.IOC) error {
	ni, ok := normalize.NormalizeIOC(i)
	if !ok {
		return nil
	}
	id := normalize.CanonicalID(ni)
	raw, err := json.Marshal(ni)
	if err != nil {
		return err
	}
	pl := ingestv1.TILinkCampaignIOCPayload{CampaignID: campaignID, IOC: raw}
	return p.publish(ctx, ingestv1.KindTILinkCampaignIOC, ingestv1.TILinkCampaignIOCIdempotencyKey(campaignID, id), pl)
}

func (p *Publisher) LinkClusterCampaign(ctx context.Context, clusterID, campaignID string) error {
	pl := ingestv1.TILinkClusterCampaignPayload{ClusterID: clusterID, CampaignID: campaignID}
	return p.publish(ctx, ingestv1.KindTILinkClusterCampaign, ingestv1.TILinkClusterCampaignIdempotencyKey(clusterID, campaignID), pl)
}

func (p *Publisher) LinkCampaignActor(ctx context.Context, campaignID, actorName string) error {
	aid := normalize.ActorStableID(actorName)
	pl := ingestv1.TILinkCampaignActorPayload{CampaignID: campaignID, ActorName: actorName}
	return p.publish(ctx, ingestv1.KindTILinkCampaignActor, ingestv1.TILinkCampaignActorIdempotencyKey(campaignID, aid), pl)
}

func (p *Publisher) LinkReportMentionsIOC(ctx context.Context, reportID string, i domain.IOC) error {
	if strings.TrimSpace(reportID) == "" {
		return nil
	}
	ni, ok := normalize.NormalizeIOC(i)
	if !ok {
		return nil
	}
	iid := normalize.CanonicalID(ni)
	raw, err := json.Marshal(ni)
	if err != nil {
		return err
	}
	pl := ingestv1.TILinkReportMentionsIOCPayload{ReportID: reportID, IOC: raw}
	return p.publish(ctx, ingestv1.KindTILinkReportMentionsIOC, ingestv1.TILinkReportMentionsIOCIdempotencyKey(reportID, iid), pl)
}

func (p *Publisher) UpsertKEVVulnerability(ctx context.Context, cve, vendor, product, summary, dateAdded string) error {
	pl := ingestv1.TIKEVVulnPayload{
		CVEID:         cve,
		VendorProject: vendor,
		Product:       product,
		ShortDesc:     summary,
		DateAdded:     dateAdded,
	}
	cveU := strings.TrimSpace(strings.ToUpper(cve))
	return p.publish(ctx, ingestv1.KindTIKEVVulnerability, ingestv1.TIKEVIdempotencyKey(cveU), pl)
}
