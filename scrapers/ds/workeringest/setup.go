package workeringest

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/threat_intelligence/pkg/ingestv1"

	neo4jstore "ds/internal/storage/neo4j"
)

// NeoConfig is Bolt credentials.
type NeoConfig struct {
	URI, Username, Password, Database string
}

// SetupDSWriter returns schema + apply + close for ingest-worker.
func SetupDSWriter(ctx context.Context, cfg NeoConfig) (
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
		return nil, nil, nil, fmt.Errorf("ds neo4j: %w", err)
	}
	ensureSchema = st.EnsureSchema
	apply = func(ctx context.Context, env *ingestv1.Envelope) error {
		return HandleDSEnvelope(ctx, st, env)
	}
	closeFn = st.Close
	return ensureSchema, apply, closeFn, nil
}
