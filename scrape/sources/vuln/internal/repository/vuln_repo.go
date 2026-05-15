package repository

import (
	"context"
	"github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln/internal/domain"
)

type VulnerabilityRepository interface {
	Save(ctx context.Context, v *domain.Vulnerability) error
	FindByCVE(ctx context.Context, id string) (*domain.Vulnerability, error)
	Upsert(ctx context.Context, v *domain.Vulnerability) error
	// MergeExploitForCVE links an Exploit node to an existing Vulnerability; no-op if CVE not in graph.
	MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error
}
