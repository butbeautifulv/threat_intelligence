package neo4j

import (
	"context"
	"fmt"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Config struct {
	URI      string
	Username string
	Password string
	Database string
}

type Client struct {
	driver   driver.DriverWithContext
	database string
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("neo4j uri is empty")
	}
	auth := driver.BasicAuth(cfg.Username, cfg.Password, "")
	d, err := driver.NewDriverWithContext(cfg.URI, auth, func(c *driver.Config) {
		c.MaxConnectionPoolSize = 50
		c.ConnectionAcquisitionTimeout = 30 * time.Second
	})
	if err != nil {
		return nil, err
	}
	if err := d.VerifyConnectivity(ctx); err != nil {
		_ = d.Close(ctx)
		return nil, err
	}
	db := cfg.Database
	if db == "" {
		db = "neo4j"
	}
	return &Client{driver: d, database: db}, nil
}

func (c *Client) Close(ctx context.Context) error { return c.driver.Close(ctx) }

func (c *Client) Session(ctx context.Context) driver.SessionWithContext {
	return c.driver.NewSession(ctx, driver.SessionConfig{
		DatabaseName: c.database,
		AccessMode:   driver.AccessModeWrite,
	})
}

func (c *Client) ExecWrite(ctx context.Context, fn func(tx driver.ManagedTransaction) error) error {
	sess := c.Session(ctx)
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx driver.ManagedTransaction) (any, error) {
		return nil, fn(tx)
	})
	return err
}

func EnsureConstraints(ctx context.Context, c *Client, queries []string) error {
	return c.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		for _, q := range queries {
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return err
			}
		}
		return nil
	})
}

