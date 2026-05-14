// Package workeringest applies ingestv1 TI messages to Neo4j (used by ingest-worker).
package workeringest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"

	"ti/internal/domain"
	tiingest "ti/internal/ingest"
	"ti/internal/repository"
	"ti/internal/usecase"
)

// HandleTIEnvelope applies a TI envelope after Validate + source checks.
func HandleTIEnvelope(ctx context.Context, repo repository.GraphRepository, uc *usecase.Ingestor, env *ingestv1.Envelope) error {
	switch env.Kind {
	case ingestv1.KindTIIoC:
		var i domain.IOC
		if err := json.Unmarshal(env.Payload, &i); err != nil {
			return err
		}
		_, err := uc.UpsertIOC(ctx, i)
		return err
	case ingestv1.KindTIKEVVulnerability:
		var p ingestv1.TIKEVVulnPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return repo.UpsertKEVVulnerability(ctx, p.CVEID, p.VendorProject, p.Product, p.ShortDesc, p.DateAdded)
	case ingestv1.KindTIReport:
		var r domain.Report
		if err := json.Unmarshal(env.Payload, &r); err != nil {
			return err
		}
		return uc.UpsertReport(ctx, r)
	case ingestv1.KindTICampaign:
		var c domain.Campaign
		if err := json.Unmarshal(env.Payload, &c); err != nil {
			return err
		}
		return uc.UpsertCampaign(ctx, c)
	case ingestv1.KindTICluster:
		var cl domain.Cluster
		if err := json.Unmarshal(env.Payload, &cl); err != nil {
			return err
		}
		return uc.UpsertCluster(ctx, cl)
	case ingestv1.KindTIActor:
		var a domain.Actor
		if err := json.Unmarshal(env.Payload, &a); err != nil {
			return err
		}
		return uc.UpsertActor(ctx, a)
	case ingestv1.KindTILinkCampaignIOC:
		var p ingestv1.TILinkCampaignIOCPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		var i domain.IOC
		if err := json.Unmarshal(p.IOC, &i); err != nil {
			return err
		}
		return repo.LinkCampaignIOC(ctx, p.CampaignID, i)
	case ingestv1.KindTILinkClusterCampaign:
		var p ingestv1.TILinkClusterCampaignPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return repo.LinkClusterCampaign(ctx, p.ClusterID, p.CampaignID)
	case ingestv1.KindTILinkCampaignActor:
		var p ingestv1.TILinkCampaignActorPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return repo.LinkCampaignActor(ctx, p.CampaignID, p.ActorName)
	case ingestv1.KindTILinkReportMentionsIOC:
		var p ingestv1.TILinkReportMentionsIOCPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		var i domain.IOC
		if err := json.Unmarshal(p.IOC, &i); err != nil {
			return err
		}
		return repo.LinkReportMentionsIOC(ctx, p.ReportID, i)
	case ingestv1.KindTIJSONLRecord:
		var p ingestv1.TIJSONLRecordPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		var lineEnv tiingest.Envelope
		if err := json.Unmarshal(p.Line, &lineEnv); err != nil {
			return fmt.Errorf("ti jsonl line: %w", err)
		}
		return uc.IngestOne(ctx, lineEnv)
	default:
		return fmt.Errorf("workeringest: unsupported TI kind %q", env.Kind)
	}
}
