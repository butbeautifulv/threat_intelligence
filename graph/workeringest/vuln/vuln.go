// Package vuln wires vuln graph ingest for ingest_worker.
package vuln

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1"

	vulningest "github.com/butbeautifulv/threat_intelligence/graph/sources/vuln/ingest"
)

// NeoConfig is Bolt credentials.
type NeoConfig = vulningest.NeoConfig

// SetupWriter returns schema + apply + close for ingest_worker.
func SetupWriter(ctx context.Context, cfg NeoConfig) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *ingestv1.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	return vulningest.SetupWriter(ctx, cfg)
}
