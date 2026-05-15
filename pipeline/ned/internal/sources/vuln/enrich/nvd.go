package enrich

import (
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	nvdmap "github.com/butbeautifulv/veil/pipeline/pkg/nvd/map"
	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/vuln/domain"
)

// FromNVDPage parses an NVD API page JSON and returns enriched vuln upsert envelopes.
func FromNVDPage(raw string) ([]*commit.Envelope, error) {
	parsed, _, err := parse.ParsePage([]byte(raw))
	if err != nil {
		return nil, err
	}
	var out []*commit.Envelope
	for _, p := range parsed {
		if strings.TrimSpace(p.CVE) == "" {
			continue
		}
		v := nvdToDomain(p)
		e, err := commit.NewEnvelope(commit.SourceVuln, commit.KindVulnUpsert, commit.VulnUpsertIdempotencyKey(v.CVE), v)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func nvdToDomain(p parse.Vulnerability) domain.Vulnerability {
	m := nvdmap.FromNVD(p)
	v := domain.Vulnerability{ID: m.ID, CVE: m.CVE, Summary: m.Summary, CWE: m.CWE}
	if len(m.CPEs) > 0 {
		v.CPEs = make([]domain.CPE, len(m.CPEs))
		for i, c := range m.CPEs {
			v.CPEs[i] = domain.CPE{URI: c.URI}
		}
	}
	if m.CVSS != nil {
		v.CVSS = &domain.CVSS{Version: m.CVSS.Version, Base: m.CVSS.Base, Vector: m.CVSS.Vector}
	}
	return v
}
