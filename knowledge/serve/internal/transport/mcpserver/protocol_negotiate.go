package mcpserver

import "encoding/json"

const (
	protocol20241105 = "2024-11-05"
	protocol20250326 = "2025-03-26"
	defaultProtocol  = protocol20241105
)

var supportedProtocols = map[string]struct{}{
	protocol20241105: {},
	protocol20250326: {},
}

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      map[string]any `json:"clientInfo"`
}

func negotiateProtocol(params json.RawMessage) string {
	if len(params) == 0 {
		return defaultProtocol
	}
	var p initializeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return defaultProtocol
	}
	if p.ProtocolVersion == "" {
		return defaultProtocol
	}
	if _, ok := supportedProtocols[p.ProtocolVersion]; ok {
		return p.ProtocolVersion
	}
	return defaultProtocol
}
