package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"

	"github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tidomain"
	"github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tiingest"
	tinormalize "github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/ti"
)

func HandleTI(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case scrapev1.KindTIKEVRow:
		var row scrapev1.TIKEVRow
		if err := json.Unmarshal(env.Payload, &row); err != nil {
			return nil, err
		}
		p := ingestv1.TIKEVVulnPayload{
			CVEID: row.CVEID, VendorProject: row.VendorProject, Product: row.Product,
			ShortDesc: row.ShortDesc, DateAdded: row.DateAdded,
		}
		cveU := strings.TrimSpace(strings.ToUpper(p.CVEID))
		out, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIKEVVulnerability, ingestv1.TIKEVIdempotencyKey(cveU), p)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindTIJSONLLine:
		var line scrapev1.TIJSONLLine
		if err := json.Unmarshal(env.Payload, &line); err != nil {
			return nil, err
		}
		pl := ingestv1.TIJSONLRecordPayload{Line: line.Line}
		var lineEnv tiingest.Envelope
		if err := json.Unmarshal(pl.Line, &lineEnv); err != nil {
			return nil, fmt.Errorf("ti jsonl: %w", err)
		}
		return tiJSONLToIngest(lineEnv)

	case scrapev1.KindTIIoCRaw:
		var i tidomain.IOC
		if err := json.Unmarshal(env.Payload, &i); err != nil {
			return nil, err
		}
		ni, ok := tinormalize.NormalizeIOC(i)
		if !ok {
			return nil, nil
		}
		id := tinormalize.CanonicalID(ni)
		out, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIIoC, ingestv1.TIIoCIdempotencyKey(id), ni)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindTIReportRaw:
		var r tidomain.Report
		if err := json.Unmarshal(env.Payload, &r); err != nil {
			return nil, err
		}
		sid := tinormalize.ReportStableID(r.Link)
		out, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIReport, ingestv1.TIReportIdempotencyKey(sid), r)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindTICampaignRaw:
		var c tidomain.Campaign
		if err := json.Unmarshal(env.Payload, &c); err != nil {
			return nil, err
		}
		c = tinormalize.NormalizeCampaign(c)
		if c.ID == "" || c.Name == "" {
			return nil, fmt.Errorf("ti campaign requires id and name")
		}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTICampaign, ingestv1.TICampaignIdempotencyKey(c.ID), c)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindTIClusterRaw:
		var cl tidomain.Cluster
		if err := json.Unmarshal(env.Payload, &cl); err != nil {
			return nil, err
		}
		cl = tinormalize.NormalizeCluster(cl)
		out, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTICluster, ingestv1.TIClusterIdempotencyKey(cl.ID), cl)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindTIActorRaw:
		var a tidomain.Actor
		if err := json.Unmarshal(env.Payload, &a); err != nil {
			return nil, err
		}
		id := strings.TrimSpace(a.ID)
		if id == "" {
			id = tinormalize.ActorStableID(a.Name)
		}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIActor, ingestv1.TIActorIdempotencyKey(id), a)
		return []*ingestv1.Envelope{out}, err

	default:
		return nil, fmt.Errorf("pipeline ti: unknown kind %q", env.Kind)
	}
}

func tiJSONLToIngest(lineEnv tiingest.Envelope) ([]*ingestv1.Envelope, error) {
	var out []*ingestv1.Envelope
	switch {
	case lineEnv.IOC != nil:
		ni, ok := tinormalize.NormalizeIOC(*lineEnv.IOC)
		if !ok {
			return nil, nil
		}
		id := tinormalize.CanonicalID(ni)
		e, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIIoC, ingestv1.TIIoCIdempotencyKey(id), ni)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Campaign != nil:
		c := tinormalize.NormalizeCampaign(*lineEnv.Campaign)
		e, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTICampaign, ingestv1.TICampaignIdempotencyKey(c.ID), c)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Cluster != nil:
		cl := tinormalize.NormalizeCluster(*lineEnv.Cluster)
		e, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTICluster, ingestv1.TIClusterIdempotencyKey(cl.ID), cl)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Actor != nil:
		a := *lineEnv.Actor
		id := strings.TrimSpace(a.ID)
		if id == "" {
			id = tinormalize.ActorStableID(a.Name)
		}
		e, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIActor, ingestv1.TIActorIdempotencyKey(id), a)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Report != nil:
		sid := tinormalize.ReportStableID(lineEnv.Report.Link)
		e, err := ingestv1.NewEnvelope(ingestv1.SourceTI, ingestv1.KindTIReport, ingestv1.TIReportIdempotencyKey(sid), *lineEnv.Report)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
