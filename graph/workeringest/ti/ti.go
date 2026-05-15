// Package ti wires TI graph ingest for ingest_worker (graph context).
package ti

import (
	"context"
	"log/slog"

	"github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1"

	tiingest "github.com/butbeautifulv/threat_intelligence/graph/sources/ti/ingest"
)

// NeoConfig is Bolt credentials for the TI writer.
type NeoConfig = tiingest.NeoConfig

// SetupWriter opens TI Neo4j and returns schema + apply + close.
func SetupWriter(ctx context.Context, cfg NeoConfig, log *slog.Logger) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *ingestv1.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	return tiingest.SetupWriter(ctx, cfg, log)
}
