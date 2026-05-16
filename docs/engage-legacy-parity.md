# Legacy MCP parity checklist

Reference: [`.external/hexstrike-ai-master/`](../.external/hexstrike-ai-master/) (MIT, not shipped).

| Area | Legacy reference | Veil engage |
|------|------------------|-------------|
| MCP tools | ~150 `@mcp.tool` | [catalog/tools.yaml](../engage/serve/catalog/tools.yaml) (150 names) |
| HTTP API | Python server :8888 | `engage-api` :8890 |
| Auth | none | Keycloak + RBAC ([pkg/auth](../pkg/auth/)) |
| Graph context | none | `ENGAGE_VEIL_API_URL` client |

Regenerate catalog:

```bash
make catalog-engage
```

Enable tools for dev in [tools.live.yaml](../engage/serve/catalog/tools.live.yaml) (`enabled: true` when binary on PATH).

## API routes (implemented)

- `GET /health`
- `GET /api/tools`, `POST /api/tools/{name}`
- `POST /api/intelligence/*` (analyze-target, select-tools, smart-scan, …)
- `POST /api/bugbounty/*` workflows
- `POST /api/visual/*` reports
- `GET /api/cache/*`, `GET /api/telemetry`
- `GET /api/processes/list|status/{pid}|dashboard`, `POST /api/processes/terminate/{pid}`
- `POST /api/jobs`, `GET /api/jobs/{id}`

## MCP (veil-engage)

- stdio: `engage/serve/cmd/mcp` (LSP framing, `tools/list`, `tools/call`)
- optional HTTP: `ENGAGE_MCP_HTTP_ENABLED=1` on `:8892`
- example: [examples/mcp/engage.stdio.json.example](../examples/mcp/engage.stdio.json.example)
