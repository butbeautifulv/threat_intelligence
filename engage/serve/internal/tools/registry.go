package tools

import (
	"fmt"
	"sort"
	"sync"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

// Registry holds tool specs by name.
type Registry struct {
	mu    sync.RWMutex
	byName map[string]tool.Spec
}

func NewRegistry(specs []tool.Spec) *Registry {
	m := make(map[string]tool.Spec, len(specs))
	for _, s := range specs {
		if s.Name == "" {
			continue
		}
		m[s.Name] = s
	}
	return &Registry{byName: m}
}

func (r *Registry) Get(name string) (tool.Spec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byName[name]
	return s, ok
}

func (r *Registry) List() []tool.Spec {
	all := r.ListAll()
	var out []tool.Spec
	for _, s := range all {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

// ListAll returns every catalog entry (including disabled) for MCP parity listing.
func (r *Registry) ListAll() []tool.Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]tool.Spec, 0, len(r.byName))
	for _, s := range r.byName {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) ListByCategory(cat toolid.Category) []tool.Spec {
	all := r.List()
	var out []tool.Spec
	for _, s := range all {
		if s.Category == cat {
			out = append(out, s)
		}
	}
	return out
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byName)
}

func (r *Registry) MustGet(name string) (tool.Spec, error) {
	s, ok := r.Get(name)
	if !ok {
		return tool.Spec{}, fmt.Errorf("unknown tool: %s", name)
	}
	if !s.Enabled {
		return tool.Spec{}, fmt.Errorf("tool disabled: %s", name)
	}
	return s, nil
}
