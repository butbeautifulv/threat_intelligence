package handle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"
)

func HandleSBOM(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case scrapev1.KindSBOMOSVJSON:
		var raw scrapev1.SBOMOSVRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := json.Unmarshal([]byte(raw.RawJSON), &doc); err != nil {
			return nil, err
		}
		affected, _ := doc["affected"].([]any)
		aff := make([]map[string]any, 0, len(affected))
		for _, a := range affected {
			if m, ok := a.(map[string]any); ok {
				aff = append(aff, m)
			}
		}
		pl := ingestv1.SBOMOSVPayload{OSVID: raw.OSVID, CVE: raw.CVE, Affected: aff}
		key := ingestv1.SBOMOSVIdempotencyKey(raw.CVE, "osv", raw.OSVID)
		out, err := ingestv1.NewEnvelope(ingestv1.SourceSBOM, ingestv1.KindSBOMOSVRecord, key, pl)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindSBOMGHSAPath:
		var raw scrapev1.SBOMGHSARaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := ingestv1.SBOMGHSAPathPayload{Path: raw.Path, Doc: raw.Doc}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceSBOM, ingestv1.KindSBOMGHSADocument, ingestv1.SBOMGHSAIdempotencyKey(raw.Path), pl)
		return []*ingestv1.Envelope{out}, err

	default:
		return nil, fmt.Errorf("pipeline sbom: unknown kind %q", env.Kind)
	}
}

func HandleCoderules(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case scrapev1.KindCoderulesCWERaw:
		var raw scrapev1.CoderulesCWERaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := ingestv1.CoderulesCWEPayload{ID: raw.ID, Name: raw.Name, Description: raw.Description, Status: raw.Status}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceCoderules, ingestv1.KindCoderulesCWERow, ingestv1.CoderulesCWEIdempotencyKey(raw.ID), pl)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindCoderulesSemgrepRaw:
		return coderulesSemgrepRaw(env.Payload)

	case scrapev1.KindCoderulesCodeQLRaw:
		return coderulesCodeQLRaw(env.Payload)

	default:
		return nil, fmt.Errorf("pipeline coderules: unknown kind %q", env.Kind)
	}
}

func HandleNuclei(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	_ = ctx
	if env.Kind != scrapev1.KindNucleiTemplateRaw {
		return nil, fmt.Errorf("pipeline nuclei: unknown kind %q", env.Kind)
	}
	return nucleiTemplateRaw(env.Payload)
}
