# MCP (Go) server for Threat Intelligence graph

Stdio **MCP** server over **Neo4j**: agents call tools to list categories, kinds, nodes, search, and neighbors—same categorical logic as the HTTP **API** ([graph/neo4jclient/query](../neo4jclient/query)).

| Field | Value |
|--------|--------|
| **Compose** | Not in default [deploy/graph/compose.yml](../../deploy/graph/compose.yml); run from source |
| **Build** | `cd graph/mcp && go build -o mcp ./cmd` |
| **Runtime doc** | [docs/threatintel-runtime.md](../../docs/threatintel-runtime.md) |
| **Depends on** | Running Neo4j (same env as `api`) |

## Run (source)

From repo root:

```bash
cd graph/mcp && go run ./cmd
```

Environment (same family as other services):

| Variable | Default |
|----------|---------|
| `NEO4J_URI` | `neo4j://localhost:7687` |
| `NEO4J_USER` | `neo4j` |
| `NEO4J_PASS` | `neo4jpassword` |
| `NEO4J_DB` | `neo4j` |

## Tools

**Category-first** (preferred for agents):

- `ti_list_categories`
- `ti_list_kinds_in_category` (`category`: e.g. `vuln`, `ti`, `sbom`, `code_rules` — see [graph/neo4jclient/query/categories.go](../neo4jclient/query/categories.go))
- `ti_nodes_by_category` (`category`, `kind`, `limit`, `offset`)
- `ti_search_in_category` (`category`, `query`, optional `kind`, `limit`)

**Legacy** (raw Neo4j labels):

- `ti_list_kinds`
- `ti_get_nodes_by_kind` (`kind`, `limit`, `offset`)
- `ti_get_node` (`id`)
- `ti_neighbors` (`id`, `depth`, `limit`)
- `ti_search` (`query`, optional `kind`, `limit`)
- `ti_health`

Implementation: shared categorical queries in [graph/neo4jclient/query](../neo4jclient/query); MCP wraps the same service as the HTTP API.
