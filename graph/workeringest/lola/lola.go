// Package lola wires lola graph ingest for ingest_worker.
package lola

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1"

	lolaingest "github.com/butbeautifulv/threat_intelligence/graph/sources/lola/ingest"
)

// NeoConfig is Bolt credentials.
type NeoConfig = lolaingest.NeoConfig

// SetupWriter returns schema + apply + close for ingest_worker.
func SetupWriter(ctx context.Context, cfg NeoConfig) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *ingestv1.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	return lolaingest.SetupWriter(ctx, cfg)
}
