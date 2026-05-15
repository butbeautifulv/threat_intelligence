package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/pkg/nvdparse"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"

	vulndomain "github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/vuln"
)

func HandleVuln(ctx context.Context, env *scrapev1.Envelope) ([]*ingestv1.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case scrapev1.KindVulnCVEUpsert:
		var v vulndomain.Vulnerability
		if err := json.Unmarshal(env.Payload, &v); err != nil {
			return nil, err
		}
		if strings.TrimSpace(v.CVE) == "" {
			return nil, fmt.Errorf("vuln: empty CVE")
		}
		out, err := ingestv1.NewEnvelope(ingestv1.SourceVuln, ingestv1.KindVulnUpsert, ingestv1.VulnUpsertIdempotencyKey(v.CVE), v)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindVulnMergeExploit:
		var raw scrapev1.VulnMergeExploit
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := ingestv1.VulnMergeExploitPayload{CVE: raw.CVE, Source: raw.Source, RefID: raw.RefID, URL: raw.URL}
		key := ingestv1.VulnMergeExploitIdempotencyKey(p.CVE, p.Source, p.RefID)
		out, err := ingestv1.NewEnvelope(ingestv1.SourceVuln, ingestv1.KindVulnMergeExploit, key, p)
		return []*ingestv1.Envelope{out}, err

	case scrapev1.KindVulnNVDPage:
		var page scrapev1.VulnNVDPage
		if err := json.Unmarshal(env.Payload, &page); err != nil {
			return nil, err
		}
		return vulnFromNVDPage(page.RawJSON)

	default:
		return nil, fmt.Errorf("pipeline vuln: unknown kind %q", env.Kind)
	}
}

func vulnFromNVDPage(raw string) ([]*ingestv1.Envelope, error) {
	parsed, _, err := nvdparse.ParsePage([]byte(raw))
	if err != nil {
		return nil, err
	}
	var out []*ingestv1.Envelope
	for _, p := range parsed {
		if strings.TrimSpace(p.CVE) == "" {
			continue
		}
		v := nvdToVulnDomain(p)
		e, err := ingestv1.NewEnvelope(ingestv1.SourceVuln, ingestv1.KindVulnUpsert, ingestv1.VulnUpsertIdempotencyKey(v.CVE), v)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func nvdToVulnDomain(p nvdparse.Vulnerability) vulndomain.Vulnerability {
	v := vulndomain.Vulnerability{
		ID:      p.ID,
		CVE:     p.CVE,
		Summary: p.Summary,
		CWE:     p.CWE,
	}
	if len(p.CPEs) > 0 {
		v.CPEs = make([]vulndomain.CPE, len(p.CPEs))
		for i, c := range p.CPEs {
			v.CPEs[i] = vulndomain.CPE{URI: c.URI}
		}
	}
	if p.CVSS != nil {
		v.CVSS = &vulndomain.CVSS{
			Version: p.CVSS.Version,
			Base:    p.CVSS.Base,
			Vector:  p.CVSS.Vector,
		}
	}
	return v
}
