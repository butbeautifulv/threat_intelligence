package mcpserver

import gq "github.com/butbeautifulv/veil/knowledge/connector/query"

type toolEntry struct {
	name        string
	description string
	schema      map[string]any
	deprecated  bool
}

func categoryEnum() []any {
	ids := gq.CategoryIDs()
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func toolDef(name, desc string, schema map[string]any) map[string]any {
	return map[string]any{
		"name":        name,
		"description": desc,
		"inputSchema": schema,
	}
}

func allToolEntries() []toolEntry {
	categoryEnum := categoryEnum()
	dep := " (deprecated; prefer category-scoped tools)"
	return []toolEntry{
		{
			name:        "ti_list_categories",
			description: "List product categories with titles and Neo4j label sets.",
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			name:        "ti_list_kinds_in_category",
			description: "List Neo4j labels within a category that exist in the graph (with counts).",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{"type": "string", "enum": categoryEnum},
				},
				"required": []string{"category"},
			},
		},
		{
			name:        "ti_nodes_by_category",
			description: "List nodes: category + kind (label must belong to that category).",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{"type": "string", "enum": categoryEnum},
					"kind":     map[string]any{"type": "string"},
					"limit":    map[string]any{"type": "integer", "default": 200},
					"offset":   map[string]any{"type": "integer", "default": 0},
				},
				"required": []string{"category", "kind"},
			},
		},
		{
			name:        "ti_search_in_category",
			description: "Search within a category (optional kind filter).",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{"type": "string", "enum": categoryEnum},
					"query":    map[string]any{"type": "string"},
					"kind":     map[string]any{"type": "string"},
					"limit":    map[string]any{"type": "integer", "default": 50},
				},
				"required": []string{"category", "query"},
			},
		},
		{
			name:        "ti_get_node",
			description: "Fetch a single node by elementId or common id fields.",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
				"required": []string{"id"},
			},
		},
		{
			name:        "ti_neighbors",
			description: "Fetch a subgraph around a node (k-hop).",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":    map[string]any{"type": "string"},
					"depth": map[string]any{"type": "integer", "default": 1, "minimum": 1, "maximum": 3},
					"limit": map[string]any{"type": "integer", "default": 500},
				},
				"required": []string{"id"},
			},
		},
		{
			name:        "ti_health",
			description: "Server and Neo4j connectivity status.",
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			name:        "ti_list_kinds",
			description: "List all distinct node labels in the graph." + dep,
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			deprecated: true,
		},
		{
			name:        "ti_get_nodes_by_kind",
			description: "List nodes for a Neo4j label without category guard." + dep,
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"kind":   map[string]any{"type": "string"},
					"limit":  map[string]any{"type": "integer", "default": 200},
					"offset": map[string]any{"type": "integer", "default": 0},
				},
				"required": []string{"kind"},
			},
			deprecated: true,
		},
		{
			name:        "ti_search",
			description: "Substring search globally or scoped to one label." + dep,
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
					"kind":  map[string]any{"type": "string"},
					"limit": map[string]any{"type": "integer", "default": 50},
				},
				"required": []string{"query"},
			},
			deprecated: true,
		},
	}
}

func listToolsPayload() map[string]any {
	entries := allToolEntries()
	tools := make([]any, len(entries))
	for i, e := range entries {
		tools[i] = toolDef(e.name, e.description, e.schema)
	}
	return map[string]any{"tools": tools}
}
