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
| engage-worker | — | Polls `ENGAGE_JOBS_DIR` when `ENGAGE_JOBS_MODE=file` (compose default) |
| engage-runner | none (profile `runner`) | Isolated toolbox image when `ENGAGE_RUNNER_MODE=docker` |

```bash
# From repo root
docker compose -f deploy/engage/compose.yml up -d --build engage-api engage-mcp engage-worker
make test-engage
make test-engage-parity
```

CI: GitHub Actions workflow [`.github/workflows/engage.yml`](../.github/workflows/engage.yml) runs `test-engage` and `test-engage-parity` on engage-related PRs.

### Runner profile (docker)

Subprocess tools run inside `engage-runner` via `docker exec` (lab only; API mounts Docker socket):

```bash
docker compose -f deploy/engage/compose.yml \
  -f deploy/engage/compose.runner.yml \
  --profile runner up -d --build engage-runner engage-api

curl -sS -X POST http://localhost:8890/api/tools/nmap_scan \
  -H 'Content-Type: application/json' \
  -d '{"target":"127.0.0.1","parameters":{"scan_type":"-sn","ports":"","additional_args":"-T4"}}'

make test-engage-smoke-tool   # opt-in; ENGAGE_SKIP_TOOL_SMOKE=1 in CI without runner
```

See [docs/engage-runtime.md](../docs/engage-runtime.md#runner-profile-docker-exec-lab-only).

### Compose e2e (async jobs)

End-to-end check: api + worker + runner (`ENGAGE_RUNNER_MODE=docker`), `POST /api/jobs` → poll until `done` or `failed`:

```bash
make test-engage-compose   # skips if docker unavailable
```

Uses `deploy/engage/compose.yml` + `compose.runner.yml` with profile `runner`. CI job `engage-compose` runs this on every engage workflow (required on GitHub Actions).

Redis job queue (prod lab):

```bash
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.queue.yml up -d redis engage-api engage-worker
# ENGAGE_JOBS_MODE=redis ENGAGE_REDIS_URL=redis://redis:6379/0
make test-engage-redis-workers   # 2 worker replicas, 10 jobs
```

NATS JetStream queue (aligns with Veil stack):

```bash
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.nats.yml up -d nats engage-api engage-worker
# ENGAGE_JOBS_MODE=nats ENGAGE_NATS_URL=nats://nats:4222
```

Browser sidecar (`--profile browser` with runner overlay):

```bash
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.runner.yml \
  --profile runner --profile browser up -d engage-browser engage-api
export ENGAGE_BROWSER_URL=http://127.0.0.1:8910
make test-engage-browser
```

## Events bus (Phase 13)

When `ENGAGE_EVENTS_NATS_ENABLED=1`, engage publishes tool audit and smart-scan findings to JetStream (`engage.events.audit`, `engage.events.finding`). The pipeline worker [engage-events](../pipeline/engage-events/) bridges to `ingest.engage.tool_run` / `ingest.engage.finding`; [graph ingest](../graph/ingest/internal/sources/engage/) persists `EngageToolRun` and `EngageFinding` in Neo4j.

```bash
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml \
  up -d --build nats engage-api engage-events-worker
make test-engage-events-pipeline

# Optional Neo4j ingest in same stack:
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml \
  --profile graph-ingest up -d ingest_worker
```

Read back via veil-api category **`engage`**: `GET /v1/categories/engage/search?q=…` (see [docs/mcp-agents.md](../docs/mcp-agents.md)).

Details: [docs/engage-runtime.md](../docs/engage-runtime.md#events-bus-e2e-nats--ingest).

## Catalog and tools

| File | Purpose |
|------|---------|
| [tools.yaml](serve/catalog/tools.yaml) | Generated list (150 tools); `make catalog-engage` — 100+ tools with structured `args` templates |
| [tools.live.yaml](serve/catalog/tools.live.yaml) | Fifteen default enabled tools for smoke / CI matrix |
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
