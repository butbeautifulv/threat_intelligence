package nvdparse

// Vulnerability is the normalized NVD CVE record (shared by scrape and pipeline).
type Vulnerability struct {
	ID      string  `json:"ID,omitempty"`
	CVE     string  `json:"CVE,omitempty"`
	Summary string  `json:"Summary,omitempty"`
	CWE     []string `json:"CWE,omitempty"`
	CPEs    []CPE   `json:"CPEs,omitempty"`
	CVSS    *CVSS   `json:"CVSS,omitempty"`
}

type CVSS struct {
	Version string  `json:"Version,omitempty"`
	Base    float64 `json:"Base,omitempty"`
	Vector  string  `json:"Vector,omitempty"`
}

type CPE struct {
	URI string `json:"URI,omitempty"`
}
