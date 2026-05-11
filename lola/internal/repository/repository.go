package repository

import (
	"context"

	"lola/internal/domain"
)

type LolaRepository interface {
	EnsureSchema(ctx context.Context) error
	UpsertArtifact(ctx context.Context, source string, a *domain.Artifact) error
}
