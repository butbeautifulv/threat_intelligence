package playbook

import (
	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

// Service exposes read-only cybersecurity playbook skills from the generated index.
type Service struct {
	cat *pbindex.Catalog
}

// NewService loads the default catalog (lazy singleton).
func NewService() (*Service, error) {
	cat, err := pbindex.Default()
	if err != nil {
		return nil, err
	}
	return &Service{cat: cat}, nil
}

func (s *Service) IndexMeta() domain.IndexFile {
	return s.cat.Meta()
}

func (s *Service) Search(query, subdomain string, limit int) []domain.SkillMeta {
	return s.cat.Search(query, subdomain, limit)
}

func (s *Service) Get(id string) (domain.SkillDetail, error) {
	return s.cat.ReadBody(id)
}

func (s *Service) ByTechnique(techniqueID string) []domain.SkillMeta {
	return s.cat.ByTechnique(techniqueID)
}

// Catalog returns the underlying index (for merge with graph query).
func (s *Service) Catalog() *pbindex.Catalog {
	return s.cat
}
