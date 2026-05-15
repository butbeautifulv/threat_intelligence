package usecase

import (
	"context"

	"github.com/butbeautifulv/threat_intelligence/graph/connector/query"
	"github.com/butbeautifulv/threat_intelligence/graph/serve/internal/domain"
)

// GraphUsecase exposes read-only graph operations for the HTTP API.
type GraphUsecase struct {
	q *query.Service
}

func NewGraphUsecase(exec query.ReadExecutor) *GraphUsecase {
	return &GraphUsecase{q: query.NewService(exec)}
}

func (u *GraphUsecase) ListCategoryMeta() []query.CategoryMeta {
	return query.ListCategoryMeta()
}

func (u *GraphUsecase) ListKindsInCategory(ctx context.Context, category string) ([]query.KindCount, error) {
	return u.q.ListKindsInCategory(ctx, category)
}

func (u *GraphUsecase) NodesByCategory(ctx context.Context, category, kind string, limit, offset int) ([]query.Node, error) {
	return u.q.NodesByCategory(ctx, category, kind, limit, offset)
}

func (u *GraphUsecase) SearchInCategory(ctx context.Context, category, q, kind string, limit int) ([]query.Node, error) {
	return u.q.SearchInCategory(ctx, category, q, kind, limit)
}

func (u *GraphUsecase) ListKinds(ctx context.Context) ([]string, error) {
	return u.q.ListKinds(ctx)
}

func (u *GraphUsecase) GetNode(ctx context.Context, id string) (*query.Node, error) {
	n, err := u.q.GetNode(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, domain.ErrNodeNotFound
	}
	return n, nil
}

func (u *GraphUsecase) Neighbors(ctx context.Context, id string, depth, limit int) (*query.Graph, error) {
	return u.q.Neighbors(ctx, id, depth, limit)
}
