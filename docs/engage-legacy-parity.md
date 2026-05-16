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

## API routes

| Route | Status |
|-------|--------|
| `GET /health` | implemented |
| `GET /api/tools`, `POST /api/tools/{name}` | implemented |
| `POST /api/intelligence/analyze-target` | implemented (HTTP/DNS heuristics + veil) |
| `POST /api/intelligence/select-tools` | implemented (`RankTools`, enabled filter; `stealth` / `comprehensive` objectives) |
| `POST /api/intelligence/create-attack-chain` | implemented (20+ `attack_patterns` + ranked fallback) |
| `POST /api/intelligence/comprehensive-api-audit` | implemented (discovery, schema, JWT, GraphQL phases) |
| `POST /api/intelligence/technology-detection` | implemented (`TechnologyStack` enum, 15 values) |
| `POST /api/intelligence/smart-scan` | implemented (`max_tools`, sync parallel or async jobs) |
| `POST /api/intelligence/assessment-report` | implemented (smart-scan + `summary_report` + findings) |
| `POST /api/intelligence/optimize-parameters` | implemented |
| `POST /api/bugbounty/*` workflows | implemented |
| `POST /api/visual/*` | implemented (`summary-report`; `export-report` → HTML/PDF) |
| `GET /api/audit/recent` | implemented (JSONL store, `ENGAGE_AUDIT_DIR`) |
| `GET /api/audit/export` | implemented (NDJSON; optional `?since=` RFC3339) |
| `GET /api/playbooks` | implemented (`playbooks/bugbounty.yaml`) |
| `POST /api/playbooks/{name}/run` | implemented (YAML → SmartScan / workflow) |
| `POST /api/intelligence/correlate-threat` | implemented (veil-graph search) |
| `POST /api/intelligence/discover-attack-chains` | implemented |
| `POST /api/intelligence/execute-attack-chain` | implemented (runs pattern steps with params) |
| `POST /api/audit/export-webhook` | implemented (`ENGAGE_AUDIT_WEBHOOK_URL`) |
| `GET /metrics` | implemented when `ENGAGE_METRICS_ENABLED=1` (Prometheus) |
| `POST /api/jobs`, `GET /api/jobs/{id}` | implemented |
| `GET /api/cache/stats`, `POST /api/cache/clear` | implemented (TTL cache) |
| `GET /api/telemetry` | implemented (uptime, jobs, processes, cache) |
| `GET /api/processes/*`, `POST terminate/pause/resume` | implemented |
| `POST /api/command` | implemented (catalog binary allowlist; `ENGAGE_ALLOW_RAW_COMMAND=1` for lab) |
| `POST /api/files/create|modify|delete`, `GET /api/files/list` | implemented (`ENGAGE_FILES_DIR`) |
| `POST /api/payloads/generate` | implemented (buffer/cyclic/random → `ENGAGE_FILES_DIR`) |
| `GET /api/jobs`, `POST /api/jobs/{id}/cancel` | implemented |
| Job backends | `ENGAGE_JOBS_MODE`: `memory`, `file`, `redis` (`ENGAGE_REDIS_URL`), `nats` (`ENGAGE_NATS_URL`, JetStream) |
| Browser tools | `ENGAGE_BROWSER_URL` sidecar — Playwright/Chromium (`--profile browser`) |
| Secure deploy | `compose.secure.yml` + nginx :8443; `make test-engage-secure` (nightly CI) |

## MCP (veil-engage)

- stdio: `engage/serve/cmd/mcp` (LSP framing, `tools/list`, `tools/call`)
- optional HTTP: `ENGAGE_MCP_HTTP_ENABLED=1` on `:8892`
- **intelligence bridge (Phase 11):** `tools/call` for `category: intelligence` and names like `comprehensive_api_audit`, `analyze_target_intelligence`, `create_attack_chain_ai` route to in-process handlers (not subprocess stubs)
- example: [examples/mcp/engage.stdio.json.example](../examples/mcp/engage.stdio.json.example)

## Phase 11 ops (optional)

| Feature | Env / path |
|---------|------------|
| Postgres audit | `ENGAGE_AUDIT_POSTGRES_URL`, `ENGAGE_AUDIT_RETENTION_DAYS` |
| Cross-layer events | `ENGAGE_EVENTS_NATS_ENABLED=1`, `ENGAGE_NATS_URL`, subjects `engage.events.audit` / `engage.events.finding`; pipeline `engage-events-worker` → `ingest.engage.tool_run` / `ingest.engage.finding`; graph ingest → Neo4j |
| Graph read (engage) | veil-api category `engage`: `GET /v1/categories/engage/search?q=`; veil-mcp search tools; `correlate-threat` returns `engage_findings` |
| PDF engine | `ENGAGE_PDF_ENGINE=gofpdf` (default) or `wkhtml` (requires `wkhtmltopdf`) |
| Playbooks | `ENGAGE_PLAYBOOKS_PATH` or `engage/serve/playbooks/bugbounty.yaml` |
| Keycloak e2e | `deploy/engage/compose.keycloak.yml`, `make test-engage-keycloak` |
| Metrics smoke | `make test-engage-metrics` |
