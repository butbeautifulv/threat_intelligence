// Package ds wires ds graph ingest for ingest_worker.
package ds

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1"

	dsingest "github.com/butbeautifulv/threat_intelligence/graph/sources/ds/ingest"
)

// NeoConfig is Bolt credentials.
type NeoConfig = dsingest.NeoConfig

// SetupWriter returns schema + apply + close for ingest_worker.
func SetupWriter(ctx context.Context, cfg NeoConfig) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *ingestv1.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	return dsingest.SetupWriter(ctx, cfg)
}
