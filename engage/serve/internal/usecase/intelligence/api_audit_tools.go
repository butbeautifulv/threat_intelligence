package intelligence

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

func (s *Service) runCatalogTool(ctx context.Context, subject, name, target string, params map[string]string) (contract.ToolRunResponse, bool) {
	if s.Tools == nil || s.Registry == nil {
		return contract.ToolRunResponse{}, false
	}
	spec, ok := s.Registry.Get(name)
	if !ok || !spec.Enabled {
		return contract.ToolRunResponse{}, false
	}
	res := s.Tools.Run(ctx, subject, name, contract.ToolRunRequest{
		Target:     target,
		Parameters: params,
	})
	return res, true
}

func countToolFindings(res contract.ToolRunResponse) int {
	if !res.Success && res.Error != "" {
		return 1
	}
	if res.Output != "" {
		return 1
	}
	return 0
}
