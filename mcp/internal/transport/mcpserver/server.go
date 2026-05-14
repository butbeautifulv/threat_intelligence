package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"

	gq "github.com/butbeautifulv/threat_intelligence/graph/query"

	"mcp/internal/usecase"
)

type Server struct {
	uc     *usecase.QueryUsecase
	logger *slog.Logger
}

func NewServer(uc *usecase.QueryUsecase, logger *slog.Logger) *Server {
	return &Server{uc: uc, logger: logger}
}

func (s *Server) Run(ctx context.Context, inReader any, outWriter any) error {
	in, ok := inReader.(interface{ Read([]byte) (int, error) })
	if !ok {
		return fmt.Errorf("invalid stdin reader")
	}
	out, ok := outWriter.(interface{ Write([]byte) (int, error) })
	if !ok {
		return fmt.Errorf("invalid stdout writer")
	}

	rw := newFramedRW(in, out)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		payload, err := rw.read(ctx)
		if err != nil {
			return err
		}
		var msg rpcMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			_ = rw.writeJSON(ctx, rpcMessage{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &rpcError{Code: -32700, Message: "Parse error"},
			})
			continue
		}

		// Notifications (no id) are ignored unless we add logging later.
		if msg.Method == "" {
			continue
		}

		resp := rpcMessage{JSONRPC: "2.0", ID: msg.ID}
		result, rerr := s.handle(ctx, msg.Method, msg.Params)
		if rerr != nil {
			resp.Error = &rpcError{Code: -32000, Message: rerr.Error()}
		} else {
			resp.Result = result
		}
		if msg.ID != nil {
			if err := rw.writeJSON(ctx, resp); err != nil {
				return err
			}
		}
	}
}

func (s *Server) handle(ctx context.Context, method string, params json.RawMessage) (any, error) {
	switch method {
	case "initialize":
		// Minimal MCP initialize response.
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]any{
				"name":    "ti-mcp-go",
				"version": "0.2.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		}, nil

	case "tools/list":
		categoryEnum := []any{"vuln", "ti", "detection", "lola", "mitre"}
		return map[string]any{
			"tools": []any{
				toolDef("ti_list_categories", "List product categories (vuln, ti, detection, lola, mitre) with titles and Neo4j label sets.", map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				}),
				toolDef("ti_list_kinds_in_category", "List Neo4j labels within a category that exist in the graph (with counts).", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category": map[string]any{"type": "string", "enum": categoryEnum},
					},
					"required": []string{"category"},
				}),
				toolDef("ti_nodes_by_category", "List nodes: category + kind (label must belong to that category).", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category": map[string]any{"type": "string", "enum": categoryEnum},
						"kind":     map[string]any{"type": "string"},
						"limit":    map[string]any{"type": "integer", "default": 200},
						"offset":   map[string]any{"type": "integer", "default": 0},
					},
					"required": []string{"category", "kind"},
				}),
				toolDef("ti_search_in_category", "Search within a category (optional kind filter).", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category": map[string]any{"type": "string", "enum": categoryEnum},
						"query":    map[string]any{"type": "string"},
						"kind":     map[string]any{"type": "string"},
						"limit":    map[string]any{"type": "integer", "default": 50},
					},
					"required": []string{"category", "query"},
				}),
				toolDef("ti_list_kinds", "List all distinct node labels in the graph (legacy; prefer ti_list_categories + ti_list_kinds_in_category).", map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				}),
				toolDef("ti_get_nodes_by_kind", "List nodes for a given Neo4j label (no category guard).", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"kind":   map[string]any{"type": "string"},
						"limit":  map[string]any{"type": "integer", "default": 200},
						"offset": map[string]any{"type": "integer", "default": 0},
					},
					"required": []string{"kind"},
				}),
				toolDef("ti_get_node", "Fetch a single node by elementId or common id fields.", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{"type": "string"},
					},
					"required": []string{"id"},
				}),
				toolDef("ti_neighbors", "Fetch a subgraph around a node (k-hop).", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":    map[string]any{"type": "string"},
						"depth": map[string]any{"type": "integer", "default": 1, "minimum": 1, "maximum": 3},
						"limit": map[string]any{"type": "integer", "default": 500},
					},
					"required": []string{"id"},
				}),
				toolDef("ti_search", "Substring search globally or scoped to one label (legacy).", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{"type": "string"},
						"kind":  map[string]any{"type": "string"},
						"limit": map[string]any{"type": "integer", "default": 50},
					},
					"required": []string{"query"},
				}),
				toolDef("ti_health", "Return basic server/runtime info.", map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				}),
			},
		}, nil

	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		return s.callTool(ctx, p.Name, p.Arguments)
	}

	return nil, fmt.Errorf("unknown method: %s", method)
}

func toolDef(name, desc string, schema map[string]any) map[string]any {
	return map[string]any{
		"name":        name,
		"description": desc,
		"inputSchema": schema,
	}
}

func (s *Server) callTool(ctx context.Context, name string, args map[string]any) (any, error) {
	switch name {
	case "ti_health":
		return toolTextResult(map[string]any{
			"ok":      true,
			"go":      runtime.Version(),
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		})

	case "ti_list_categories":
		return toolTextResult(map[string]any{"categories": gq.ListCategoryMeta()})

	case "ti_list_kinds_in_category":
		cat := getString(args, "category")
		kinds, err := s.uc.ListKindsInCategory(ctx, cat)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"category": cat, "kinds": kinds})

	case "ti_nodes_by_category":
		cat := getString(args, "category")
		kind := getString(args, "kind")
		limit := getInt(args, "limit", 200)
		offset := getInt(args, "offset", 0)
		nodes, err := s.uc.NodesByCategory(ctx, cat, kind, limit, offset)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"category": cat, "kind": kind, "nodes": nodes})

	case "ti_search_in_category":
		cat := getString(args, "category")
		q := getString(args, "query")
		kind := getString(args, "kind")
		limit := getInt(args, "limit", 50)
		nodes, err := s.uc.SearchInCategory(ctx, cat, q, kind, limit)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"category": cat, "query": q, "kind": kind, "nodes": nodes})

	case "ti_list_kinds":
		kinds, err := s.uc.ListKinds(ctx)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"kinds": kinds})

	case "ti_get_nodes_by_kind":
		kind := getString(args, "kind")
		limit := getInt(args, "limit", 200)
		offset := getInt(args, "offset", 0)
		nodes, err := s.uc.NodesByKind(ctx, kind, limit, offset)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"nodes": nodes})

	case "ti_get_node":
		id := getString(args, "id")
		n, err := s.uc.GetNode(ctx, id)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"node": n})

	case "ti_neighbors":
		id := getString(args, "id")
		depth := getInt(args, "depth", 1)
		limit := getInt(args, "limit", 500)
		g, err := s.uc.Neighbors(ctx, id, depth, limit)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"graph": g})

	case "ti_search":
		q := getString(args, "query")
		kind := getString(args, "kind")
		limit := getInt(args, "limit", 50)
		nodes, err := s.uc.Search(ctx, q, kind, limit)
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"nodes": nodes})
	}

	s.logger.Warn("unknown tool", slog.String("name", name))
	return nil, fmt.Errorf("unknown tool: %s", name)
}

func toolTextResult(v any) (any, error) {
	// MCP tool results commonly return content array.
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"content": []any{
			map[string]any{"type": "text", "text": string(b)},
		},
	}, nil
}

func getString(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]any, k string, def int) int {
	if m == nil {
		return def
	}
	v, ok := m[k]
	if !ok {
		return def
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return def
	}
}

