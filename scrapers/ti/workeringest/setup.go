package workeringest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"

	neo4jstore "ti/internal/storage/neo4j"
	"ti/internal/usecase"
)

// NeoConfig is Bolt credentials for the TI Neo4j writer (same shape as other scrapers).
type NeoConfig struct {
	URI, Username, Password, Database string
}

// SetupTIWriter opens the TI graph store and returns schema + apply + close for ingest-worker.
func SetupTIWriter(ctx context.Context, cfg NeoConfig, log *slog.Logger) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *ingestv1.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	st, err := neo4jstore.New(ctx, neo4jstore.Config{
		URI:      cfg.URI,
		Username: cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ti neo4j: %w", err)
	}
	uc := usecase.NewIngestor(st, log)
	ensureSchema = st.EnsureSchema
	apply = func(ctx context.Context, env *ingestv1.Envelope) error {
		return HandleTIEnvelope(ctx, st, uc, env)
	}
	closeFn = st.Close
	return ensureSchema, apply, closeFn, nil
}
