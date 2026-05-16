# MCP for agents (Veil)

Veil exposes **two MCP servers** for agents — keep them separate:

| MCP | Layer | Purpose |
|-----|-------|---------|
| **veil-mcp** | Graph read | Query Neo4j TI data (categories, nodes, search) |
| **veil-engage** | Engage exec | Run security tools from YAML catalog (~150 names) |

Do not merge offensive tool execution into the graph MCP process.

---

## veil-mcp (graph read)

The **veil-mcp** server exposes read-only threat-intel graph tools over [MCP](https://modelcontextprotocol.io/) **stdio** (LSP-framed JSON-RPC) and optional **Streamable HTTP**. It works with Cursor, Claude Desktop, VS Code Copilot, Roo Code, Cline, and any MCP-compatible client.

## Stdio transport (important)

MCP on stdio uses **stdout only for JSON-RPC** (`Content-Length` framing). All logs go to **stderr** (`SetupMCPLogger`). Do not redirect stderr to `/dev/null` when debugging — only stdout must stay clean for the client.

Recommended launcher (builds `bin/mcp` if missing):

```bash
./scripts/mcp/run-veil-mcp.sh
```

## Prerequisites

1. Neo4j with graph data (compose bootstrap or ingest pipeline).
2. Binary (or use the launcher above):

```bash
cd graph/serve && env GOWORK=../go.work go build -o bin/mcp ./cmd/mcp
```

Default Bolt: `neo4j://localhost:7687`, user `neo4j`, password `neo4jpassword`.

## Protocol versions

Supported: `2024-11-05`, `2025-03-26`. On `initialize`, the server **echoes** the client's version when it is in that set; otherwise it returns `2024-11-05`. Streamable HTTP clients that omit a version may receive `2025-03-26`.

## Agent compatibility matrix

| Client | Config location | Schema | Example |
|--------|-----------------|--------|---------|
| **Cursor** | Project `.cursor/mcp.json` or Settings → MCP | `mcpServers` + `command` | [cursor.mcp.json.example](../examples/mcp/cursor.mcp.json.example) |
| **Claude Desktop** | `claude_desktop_config.json` | `mcpServers` | [claude-desktop.mcp.json.example](../examples/mcp/claude-desktop.mcp.json.example) |
| **VS Code Copilot** | `.vscode/mcp.json` | `servers` + `type: "stdio"` | [vscode-copilot.mcp.json.example](../examples/mcp/vscode-copilot.mcp.json.example) |
| **Roo Code / Cline** | Same as Cursor | `mcpServers` | [cursor.mcp.json.example](../examples/mcp/cursor.mcp.json.example) |
| **Generic stdio** | Varies | `mcpServers` | [mcp.json.example](../examples/mcp/mcp.json.example) |
| **HTTP MCP** | Client-dependent | `url` + `headers` | [mcp.remote.json.example](../examples/mcp/mcp.remote.json.example) |
| **Auth (Keycloak)** | stdio `env` | `MCP_ACCESS_TOKEN` | [mcp.auth.json.example](../examples/mcp/mcp.auth.json.example) |

Common fields (recommended):

- `command`: path to [scripts/mcp/run-veil-mcp.sh](../scripts/mcp/run-veil-mcp.sh)
- `timeout`: `300` (seconds) for heavy Neo4j queries
- `description`: note that tools are **read-only** — do not use `alwaysAllow` as for offensive automation MCPs

### Cursor

Copy [examples/mcp/cursor.mcp.json.example](../examples/mcp/cursor.mcp.json.example), fix the `command` path, reload MCP in Settings.

### Claude Desktop

macOS: `~/Library/Application Support/Claude/claude_desktop_config.json` — merge the `mcpServers` block from [claude-desktop.mcp.json.example](../examples/mcp/claude-desktop.mcp.json.example).

### VS Code Copilot

Use [vscode-copilot.mcp.json.example](../examples/mcp/vscode-copilot.mcp.json.example) in `.vscode/mcp.json` (`servers`, not `mcpServers`).

### Manual verification checklist

| Step | Cursor | Claude Desktop | VS Code Copilot |
|------|--------|----------------|-----------------|
| Server shows as connected | Expected | Expected | Expected |
| `tools/list` includes `ti_list_categories` | Expected | Expected | Expected |
| `ti_health` returns `neo4j_ok: true` | Expected | Expected | Expected |
| Automated in CI | — | — | — |

CI covers stdio via `./scripts/smoke/mcp-smoke.sh` and unit tests via `make test-graph-serve`.

## Environment

| Variable | Default | Role |
|----------|---------|------|
| `NEO4J_URI` | `neo4j://localhost:7687` | Bolt URI |
| `NEO4J_USER` | `neo4j` | Neo4j user |
| `NEO4J_PASS` | `neo4jpassword` | Neo4j password |
| `NEO4J_DB` | `neo4j` | Database name |
| `MCP_ENV` | `local` | Log level (`local` / `dev` / `prod`) — logs on **stderr** |
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

## Docker (stdio)

```bash
docker build -f deploy/graph/docker/mcp.Dockerfile -t veil-mcp .
docker run -i --rm --network host \
  -e NEO4J_URI=neo4j://localhost:7687 \
  -e NEO4J_USER=neo4j -e NEO4J_PASS=neo4jpassword \
  veil-mcp
```

## Tools (parity with HTTP API)

| Tool | HTTP equivalent |
|------|-----------------|
| `ti_list_categories` | `GET /v1/categories` |
| `ti_list_kinds_in_category` | `GET /v1/categories/{category}/kinds` |
| `ti_nodes_by_category` | `GET /v1/categories/{category}/nodes` |
| `ti_search_in_category` | `GET /v1/categories/{category}/search` |
| `ti_get_node` | `GET /v1/nodes/{id}` |
| `ti_neighbors` | `GET /v1/nodes/{id}/neighbors` |
| `ti_health` | connectivity + runtime |
| `ping` | MCP keepalive (empty result) |

Legacy (deprecated): `ti_list_kinds`, `ti_get_nodes_by_kind`, `ti_search`.

## Smoke test

```bash
./scripts/smoke/mcp-smoke.sh
```

Uses protocol `2024-11-05` (typical for Cursor/Claude clients).

## Authentication (optional)

When `AUTH_ENABLED=1`:

- Set `KEYCLOAK_ISSUER`, `RBAC_*` as in [auth-keycloak.md](auth-keycloak.md).
- Put a valid access token in `MCP_ACCESS_TOKEN` in the MCP server `env` block.
- `initialize` / `tools/list` stay open on stdio; **`tools/call`** requires a valid token and role.

## MCP Streamable HTTP/SSE

```bash
export MCP_HTTP_ENABLED=1
./scripts/mcp/run-veil-mcp.sh
```

Endpoint: `POST http://localhost:8091/mcp`, `GET /health`.

Remote client: [mcp.remote.json.example](../examples/mcp/mcp.remote.json.example).

## veil-engage (tool execution)

Engage MCP runs separately from graph read:

```bash
./scripts/mcp/run-veil-engage.sh
```

| Client | Example |
|--------|---------|
| Cursor / Cline | [engage.stdio.json.example](../examples/mcp/engage.stdio.json.example) |
| HTTP MCP | [engage.http.json.example](../examples/mcp/engage.http.json.example) |

- Server name: `veil-engage`
- Methods: `initialize`, `tools/list` (~150 catalog tools), `tools/call` → `POST /api/tools/{name}` equivalent
- Auth: `AuthorizeEngageMCP` + role `veil-engage-runner` when `AUTH_ENABLED=1`
- Logs on **stderr** (same stdio rule as veil-mcp)

Compose: `deploy/engage/compose.yml` (`engage-mcp` on :8892). Docs: [engage-runtime.md](engage-runtime.md), [engage-legacy-parity.md](engage-legacy-parity.md).

### Cross-layer workflow (engage scan → graph read)

When `ENGAGE_EVENTS_NATS_ENABLED=1` and the events bus is running, tool runs and findings are ingested into Neo4j as `EngageToolRun` / `EngageFinding` nodes (category **`engage`**).

1. Run a scan with **veil-engage** (`httpx_probe`, `smart-scan`, etc.).
2. Query results with **veil-mcp** or veil-api: `GET /v1/categories/engage/search?q=example.com`.
3. Optional: `correlate_threat_intelligence` on engage-api merges TI/vuln/engage hits when `ENGAGE_VEIL_API_URL` is set.

Smoke: `make test-engage-events-pipeline` (Docker, includes Neo4j assert with `--profile graph-ingest`).

## Related

- [engage-runtime.md](engage-runtime.md) — engage API, runner modes, ports
- [external-hexstrike.md](external-hexstrike.md) — MIT reference in `.external/` (superseded by engage layer)
- [auth-keycloak.md](auth-keycloak.md) — Keycloak, RBAC
- [deploy-secure.md](deploy-secure.md) — production hardening
- [threatintel-runtime.md](threatintel-runtime.md) — compose, ports
