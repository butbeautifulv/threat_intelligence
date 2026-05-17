package mcpserver

import (
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
)

func listToolsPayload(specs []tool.Spec) map[string]any {
	tools := make([]map[string]any, 0, len(specs))
	for _, s := range specs {
		desc := s.Description
		if !s.Enabled {
			desc = desc + " (disabled until enabled in catalog)"
		}
		tools = append(tools, map[string]any{
			"name":        s.Name,
			"description": desc,
			"inputSchema": s.InputSchema(),
		})
	}
	return map[string]any{"tools": tools}
}
