# MCP (Go) server for Threat Intelligence graph

This service exposes a small MCP tool surface over **Neo4j**, so an agent can query node categories (labels) and fetch nodes/subgraphs via tool calls.

## Run

From repo root:

```bash
cd mcp
go run ./cmd
```

Env (same as other services):

- `NEO4J_URI` (default `neo4j://localhost:7687`)
- `NEO4J_USER` (default `neo4j`)
- `NEO4J_PASS` (default `neo4jpassword`)
- `NEO4J_DB` (default `neo4j`)

## Tools

Category-first (preferred for agents):

- `ti_list_categories`
- `ti_list_kinds_in_category` (`category`: any key from `GET /v1/categories` / [graph/query/categories.go](../graph/query/categories.go), e.g. `vuln`, `ti`, `sbom`, `code_rules`)
- `ti_nodes_by_category` (`category`, `kind`, `limit`, `offset`)
- `ti_search_in_category` (`category`, `query`, optional `kind`, `limit`)

Legacy (raw Neo4j labels):

- `ti_list_kinds`
- `ti_get_nodes_by_kind` (`kind`, `limit`, `offset`)
- `ti_get_node` (`id`)
- `ti_neighbors` (`id`, `depth`, `limit`)
- `ti_search` (`query`, optional `kind`, `limit`)
- `ti_health`

Implementation: shared categorical queries live in [graph/query](../graph/query); MCP wraps the same service as the HTTP API.

