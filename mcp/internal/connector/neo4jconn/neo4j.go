package neo4jconn

import (
	"context"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	graphneo4j "github.com/butbeautifulv/threat_intelligence/graph/neo4j"
)

type Config = graphneo4j.Config

type Connector struct {
	client *graphneo4j.Client
}

func New(ctx context.Context, cfg Config) (*Connector, error) {
	c, err := graphneo4j.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Connector{client: c}, nil
}

func (c *Connector) Close(ctx context.Context) error { return c.client.Close(ctx) }

func (c *Connector) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	sess := c.client.Session(ctx)
	defer sess.Close(ctx)
	return sess.ExecuteRead(ctx, fn)
}

