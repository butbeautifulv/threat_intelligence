package components

import (
	"context"

	"github.com/butbeautifulv/veil/graph/serve/internal/config"
	neo4jstore "github.com/butbeautifulv/veil/graph/serve/internal/storage/neo4j"
	"github.com/butbeautifulv/veil/graph/serve/internal/usecase"
)

type APIComponents struct {
	Neo4jStore *neo4jstore.Store
	Graph      *usecase.GraphUsecase
}

func InitAPI(cfg *config.Config) (*APIComponents, error) {
	store, err := neo4jstore.New(context.Background(), neo4jstore.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
		Database: cfg.Neo4j.Database,
	})
	if err != nil {
		return nil, err
	}
	return &APIComponents{
		Neo4jStore: store,
		Graph:      usecase.NewGraphUsecase(store),
	}, nil
}

func (c *APIComponents) Shutdown() {
	if c.Neo4jStore != nil {
		_ = c.Neo4jStore.Close(context.Background())
	}
}
