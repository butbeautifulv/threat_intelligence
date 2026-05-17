package mcpserver

import (
	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

func listToolsPayload(specs []tool.Spec) map[string]any {
	tools := make([]mcp.ToolDescriptor, 0, len(specs))
	for _, s := range specs {
		desc := s.Description
		if !s.Enabled {
			desc = desc + " (disabled until enabled in catalog)"
		}
		tools = append(tools, mcp.ToolDescriptor{
			Name:        s.Name,
			Description: desc,
			InputSchema: s.InputSchema(),
		})
	}
	return mcp.ListToolsPayload(tools)
}
