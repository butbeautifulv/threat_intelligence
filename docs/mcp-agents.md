# MCP for agents (Veil graph read)

The **veil-mcp** server exposes read-only threat-intel graph tools over [MCP](https://modelcontextprotocol.io/) stdio. It works with any MCP client (Cursor, Claude Desktop, CLI, SDK agents) — not tied to a specific IDE.

## Prerequisites

1. Neo4j with graph data (compose bootstrap or ingest pipeline).
2. Built binary:

```bash
cd graph/serve && env GOWORK=../go.work go build -o bin/mcp ./cmd/mcp
```

Default Bolt: `neo4j://localhost:7687`, user `neo4j`, password `neo4jpassword`.

## Environment

| Variable | Default | Role |
|----------|---------|------|
| `NEO4J_URI` | `neo4j://localhost:7687` | Bolt URI |
| `NEO4J_USER` | `neo4j` | Neo4j user |
| `NEO4J_PASS` | `neo4jpassword` | Neo4j password |
| `NEO4J_DB` | `neo4j` | Database name |
| `MCP_ENV` | `local` | Log level (`local` / `dev` / `prod`) |
| `APP_VERSION` / `MCP_VERSION` | `0.4.2` | Reported server version |
| `AUTH_ENABLED` | `0` | Require Keycloak JWT for `tools/call` |
| `RBAC_ENABLED` | `0` | Require `veil-reader` / `veil-admin` roles |
| `KEYCLOAK_ISSUER` | — | Realm issuer (required if auth on) |
| `MCP_ACCESS_TOKEN` | — | Access token for stdio MCP when auth on |
| `MCP_HTTP_ENABLED` | `0` | Streamable HTTP on `MCP_HTTP_LISTEN` (default `:8091`) |
| `MCP_HTTP_LISTEN` | `:8091` | HTTP listen address |
| `MCP_HTTP_PATH` | `/mcp` | MCP endpoint path |
| `MCP_HTTP_PREFER_SSE` | `0` | Prefer SSE responses on POST |
| `MCP_HTTP_BIND_LOCAL` | `0` | Bind HTTP to `127.0.0.1` only |

Full auth setup: [auth-keycloak.md](auth-keycloak.md).

## Client configuration

Copy [examples/mcp/mcp.json.example](../examples/mcp/mcp.json.example) and set `command` to your `bin/mcp` path.

### Cursor

Settings → MCP → add server, or project/user `mcp.json` with the same `command` + `env` block.

### Claude Desktop

`~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or platform equivalent — same `mcpServers` shape as the example.

### Docker (stdio)

Build image from repo root:

```bash
docker build -f deploy/graph/docker/mcp.Dockerfile -t veil-mcp .
```

Run interactively against host Neo4j:

```bash
docker run -i --rm --network host \
  -e NEO4J_URI=neo4j://localhost:7687 \
  -e NEO4J_USER=neo4j -e NEO4J_PASS=neo4jpassword \
  veil-mcp
```

Point the MCP client at `docker run -i ...` as the command (wrapper script recommended).

## Tools (parity with HTTP API)

Categorical tools match [openapi.yaml](openapi.yaml):

| Tool | HTTP equivalent |
|------|-----------------|
| `ti_list_categories` | `GET /v1/categories` |
| `ti_list_kinds_in_category` | `GET /v1/categories/{category}/kinds` |
| `ti_nodes_by_category` | `GET /v1/categories/{category}/nodes` |
| `ti_search_in_category` | `GET /v1/categories/{category}/search` |
| `ti_get_node` | `GET /v1/nodes/{id}` |
| `ti_neighbors` | `GET /v1/nodes/{id}/neighbors` |
| `ti_health` | connectivity + runtime (richer than `GET /health`) |

Legacy (deprecated, logged on use): `ti_list_kinds`, `ti_get_nodes_by_kind`, `ti_search`.

## Smoke test

With graph stack up:

```bash
./scripts/smoke/mcp-smoke.sh
```

## Authentication (optional)

When `AUTH_ENABLED=1`:

- Set `KEYCLOAK_ISSUER`, `RBAC_*` as in [auth-keycloak.md](auth-keycloak.md).
- Put a valid access token in `MCP_ACCESS_TOKEN` in the MCP server `env` block.
- `tools/list` and `initialize` stay open; **`tools/call`** returns `unauthorized` / `forbidden` without a valid token and role.

## MCP Streamable HTTP/SSE

Enable remote MCP clients (same `cmd/mcp` binary; stdio keeps running):

```bash
export MCP_HTTP_ENABLED=1
export MCP_HTTP_LISTEN=:8091
cd graph/serve && go run ./cmd/mcp
```

Endpoint: `POST http://localhost:8091/mcp` (also `GET /health` on the same port).

### curl example

```bash
curl -sS -X POST http://localhost:8091/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}'
```

With Keycloak (`AUTH_ENABLED=1`):

```bash
curl -sS -X POST http://localhost:8091/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ti_health","arguments":{}}}'
```

### SSE responses

If the client sends `Accept: text/event-stream` (or `MCP_HTTP_PREFER_SSE=1`), responses use Server-Sent Events (`event: message`, `data: <json-rpc>`).

`GET /mcp` returns **405** (no standalone listener in v1). JSON-RPC **batch** is not supported.

### Remote client config

See [examples/mcp/mcp.remote.json.example](../examples/mcp/mcp.remote.json.example) for URL + Bearer header (clients that support HTTP MCP transport).

## Related

- [auth-keycloak.md](auth-keycloak.md) — Keycloak, AD federation, RBAC roles
- [threatintel-runtime.md](threatintel-runtime.md) — compose, API port 8090
- [graph/serve/](../graph/serve/) — module layout
