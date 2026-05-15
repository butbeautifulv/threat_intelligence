// Package nvdmap converts NVD API vulnerabilities into the shared canonical shape.
package nvdmap

import "github.com/butbeautifulv/threat_intelligence/pipeline/pkg/nvd/parse"

// Vulnerability is the canonical NVD-derived vulnerability shape shared across layers.
type Vulnerability struct {
	ID      string   `json:"ID,omitempty"`
	CVE     string   `json:"CVE,omitempty"`
	Summary string   `json:"Summary,omitempty"`
	CWE     []string `json:"CWE,omitempty"`
	CPEs    []CPE    `json:"CPEs,omitempty"`
	CVSS    *CVSS    `json:"CVSS,omitempty"`
}

type CVSS struct {
	Version string  `json:"Version,omitempty"`
	Base    float64 `json:"Base,omitempty"`
	Vector  string  `json:"Vector,omitempty"`
}

type CPE struct {
	URI string `json:"URI,omitempty"`
}

// FromNVD maps an NVD API vulnerability into the shared shape.
func FromNVD(p parse.Vulnerability) Vulnerability {
	v := Vulnerability{
		ID:      p.ID,
		CVE:     p.CVE,
		Summary: p.Summary,
		CWE:     p.CWE,
	}
	if len(p.CPEs) > 0 {
		v.CPEs = make([]CPE, len(p.CPEs))
		for i, c := range p.CPEs {
			v.CPEs[i] = CPE{URI: c.URI}
		}
	}
	if p.CVSS != nil {
		v.CVSS = &CVSS{
			Version: p.CVSS.Version,
			Base:    p.CVSS.Base,
			Vector:  p.CVSS.Vector,
		}
	}
	return v
}
