# MCP (Go) server for Threat Intelligence graph

Stdio **MCP** server over **Neo4j**: agents call tools to list categories, kinds, nodes, search, and neighbors—same categorical logic as the HTTP **API** ([graph/query](../graph/query)).

| Field | Value |
|--------|--------|
| **Compose service** | `mcp` (profile **`mcp`**) |
| **Build** | [docker/mcp.Dockerfile](../docker/mcp.Dockerfile) |
| **Runtime doc** | [docs/threatintel-runtime.md](../docs/threatintel-runtime.md#mcp-stdio) |
| **Depends on** | Healthy Neo4j + completed `graph-bootstrap` when using Compose |

---

## Run (source)

From repo root:

```bash
cd mcp
go run ./cmd
```

Environment (same family as other services):

| Variable | Default |
|----------|---------|
| `NEO4J_URI` | `neo4j://localhost:7687` |
| `NEO4J_USER` | `neo4j` |
| `NEO4J_PASS` | `neo4jpassword` |
| `NEO4J_DB` | `neo4j` |

---

## Run (Docker Compose)

After the default stack is up (`neo4j` + `graph-bootstrap` + `api`):

```bash
docker compose --profile mcp run --rm -i mcp
```

---

## Tools

**Category-first** (preferred for agents):

- `ti_list_categories`
- `ti_list_kinds_in_category` (`category`: any key from `GET /v1/categories` / [graph/query/categories.go](../graph/query/categories.go), e.g. `vuln`, `ti`, `sbom`, `code_rules`)
- `ti_nodes_by_category` (`category`, `kind`, `limit`, `offset`)
- `ti_search_in_category` (`category`, `query`, optional `kind`, `limit`)

**Legacy** (raw Neo4j labels):

- `ti_list_kinds`
- `ti_get_nodes_by_kind` (`kind`, `limit`, `offset`)
- `ti_get_node` (`id`)
- `ti_neighbors` (`id`, `depth`, `limit`)
- `ti_search` (`query`, optional `kind`, `limit`)
- `ti_health`

Implementation: shared categorical queries live in [graph/query](../graph/query); MCP wraps the same service as the HTTP API.
