# Engage layer (active security testing)

Fourth Veil runtime context: **authorized tool execution**, intelligence workflows, and structured reports. Threat-intel **read** stays in [graph/serve](../graph/serve/) (`veil-mcp`); **execution** is here (`veil-engage`).

## What it is

Greenfield **Go** implementation of the tool-orchestration model from the MIT reference in [`.external/hexstrike-ai-master/`](../.external/hexstrike-ai-master/) (attribution: [NOTICE.hexstrike](NOTICE.hexstrike)). Veil does **not** ship or run that Python stack — engage provides:

- **YAML catalog** — 150 legacy MCP tool names with per-tool `parameters` and `args` templates
- **Generic runner** — subprocess execution with Keycloak RBAC and audit logging
- **HTTP API** — unified `POST /api/tools/{name}` (not 90 separate Flask routes)
- **MCP** — `tools/list` and `tools/call` for Cursor, Claude Desktop, VS Code Copilot, etc.
- **Optional graph context** — service-account JWT to `veil-api` for TI enrichment

## Layout

| Module | Path |
|--------|------|
| **serve** | [serve/](serve/) — `engage-api`, `veil-engage` MCP, `engage-worker` |
| **catalog** | [serve/catalog/](serve/catalog/) — `tools.yaml`, `tools.live.yaml`, `tools.enabled.yaml` |
| **pkg** | [pkg/engage/](../pkg/engage/) — contracts, tool categories |

## Services (dev compose)

| Service | Port / transport | Role |
|---------|------------------|------|
| engage-api | :8890 | REST: tools, intelligence, bugbounty workflows, jobs, processes |
| engage-mcp | stdio or :8892 | MCP for agents |
| engage-worker | — | Async job queue (in-process; shares catalog with API) |
| engage-runner | none (profile `runner`) | Isolated toolbox image when `ENGAGE_RUNNER_MODE=docker` |

```bash
# From repo root
docker compose -f deploy/engage/compose.yml up -d --build engage-api engage-mcp
make test-engage
make test-engage-parity
```

## Catalog and tools

| File | Purpose |
|------|---------|
| [tools.yaml](serve/catalog/tools.yaml) | Generated list (150 tools); `make catalog-engage` |
| [tools.live.yaml](serve/catalog/tools.live.yaml) | Five default enabled tools for smoke |
| [tools.enabled.yaml](serve/catalog/tools.enabled.yaml) | Overrides from [enable-catalog-by-category.sh](../scripts/engage/enable-catalog-by-category.sh) |

Example — nmap with parameters:

```bash
curl -sS -X POST http://localhost:8890/api/tools/nmap_scan \
  -H 'Content-Type: application/json' \
  -d '{"target":"scanme.nmap.org","parameters":{"scan_type":"-sV","ports":"80,443"}}'
```

## MCP (veil-engage)

```bash
./scripts/mcp/run-veil-engage.sh
```

Examples: [engage.stdio.json.example](../examples/mcp/engage.stdio.json.example), [engage.http.json.example](../examples/mcp/engage.http.json.example).

## Boundaries

- **Does not** import `scrape/`, `pipeline/`, or `graph/ingest`
- **Does not** connect to Neo4j directly — use `ENGAGE_VEIL_API_URL` → veil-api
- **May** import `pkg/auth`, `pkg/engage/*`

## Docs

- [docs/engage-runtime.md](../docs/engage-runtime.md) — env, ports, threat model, runner modes
- [docs/engage-tools.md](../docs/engage-tools.md) — catalog schema
- [docs/engage-legacy-parity.md](../docs/engage-legacy-parity.md) — route and MCP parity matrix
- [docs/external-hexstrike.md](../docs/external-hexstrike.md) — reference-only `.external/` tree
