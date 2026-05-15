package domain

// AdvisoryRef points at an OSV/GHSA document for a CVE.
type AdvisoryRef struct {
	CVE    string
	Source string
	Path   string
}
