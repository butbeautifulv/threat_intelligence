package veilgraph

import (
	"context"
	"encoding/json"
)

// Reader is the veil-api read surface used by engage intelligence.
type Reader interface {
	Enabled() bool
	Categories(ctx context.Context) (json.RawMessage, error)
	Search(ctx context.Context, category, query string) (json.RawMessage, error)
	EngageContext(ctx context.Context, host string) (json.RawMessage, error)
	GetNode(ctx context.Context, id string) (json.RawMessage, error)
	Neighbors(ctx context.Context, id string, depth int) (json.RawMessage, error)
	PlaybooksByTechnique(ctx context.Context, techniqueID string) (json.RawMessage, error)
	PlaybookRecommendTools(ctx context.Context, skillID, techniqueID string) (json.RawMessage, error)
}
