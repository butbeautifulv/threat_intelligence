// Package normalize forwards to pkg/ti/normalize (SOT). Prefer importing github.com/butbeautifulv/veil/pkg/ti/normalize in new code.
package normalize

import pkgn "github.com/butbeautifulv/veil/pkg/ti/normalize"

var (
	NormalizeIOC      = pkgn.NormalizeIOC
	NormalizeCampaign = pkgn.NormalizeCampaign
	NormalizeCluster  = pkgn.NormalizeCluster
	CanonicalID       = pkgn.CanonicalID
	ActorStableID     = pkgn.ActorStableID
	ReportStableID    = pkgn.ReportStableID
)
