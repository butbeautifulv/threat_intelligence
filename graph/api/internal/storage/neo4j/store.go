package neo4j

import (
	"context"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	graphneo4j "github.com/butbeautifulv/threat_intelligence/graph/neo4jclient/neo4j"
)

type Config = graphneo4j.Config

// Store is the Neo4j read connector for the API (implements graph/query.ReadExecutor).
type Store struct {
	client *graphneo4j.Client
}

func New(ctx context.Context, cfg Config) (*Store, error) {
	c, err := graphneo4j.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{client: c}, nil
}

func (s *Store) Close(ctx context.Context) error { return s.client.Close(ctx) }

func (s *Store) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	return s.client.ExecRead(ctx, fn)
}
