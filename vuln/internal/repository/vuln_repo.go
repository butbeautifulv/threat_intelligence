package repository

import (
	"context"
	"vuln/internal/domain"
)

type VulnerabilityRepository interface {
	Save(ctx context.Context, v *domain.Vulnerability) error
	FindByCVE(ctx context.Context, id string) (*domain.Vulnerability, error)
	Upsert(ctx context.Context, v *domain.Vulnerability) error
}
