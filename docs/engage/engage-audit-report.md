# Engage migration audit report (HexStrike → Veil)

**Date:** 2026-05-16  
**Scope:** R0–R120 / Phase 16–23 closure verification per [hexstrike migration audit plan](../.cursor/plans/archive/hexstrike_migration_audit_12c9842f.plan.md).

## Verdict

| Dimension | Status | Notes |
|-----------|--------|-------|
| **Architecture** | **Confirmed** | Go engage layer replaces Python monolith: catalog + unified `POST /api/tools/{name}`, Keycloak, events→Neo4j, veil read |
| **MCP / catalog names** | **OK** | 151 legacy `@mcp.tool` + 8 engage bridge tools → **158** catalog entries |
| **HTTP route parity** | **OK** | 156 legacy routes: 59 implemented, 90 N/A (unified tools), 7 N/A (out of scope); **0 unexplained missing** — see [engage-route-parity.csv](engage-route-parity.csv) |
| **Execution breadth** | **OK** (release gates) | **158/158** catalog; **54** bridge (`make test-engage-bridge-coverage`); **104** subprocess enabled (`make test-engage-na-matrix`, 0 `runner_N/A`). Callable matrix: `make test-engage-executable-matrix` |
| **README KPI (24×, 98.7%)** | **Not claimed** | Benchmark script regression-only |

## Automated gates

| Gate | Result | Notes |
|------|--------|-------|
| `make test-engage` | **PASS** | Unit tests + api/mcp/worker build |
| `make test-engage-parity` | **PASS** | 151 MCP + 8 bridge tools in catalog |
| `make test-engage-catalog-args` | **PASS** | 158 tools; 112 non-generic args; 60 documented generic |
| `make test-engage-decision-parity` | **PASS** | Effectiveness tables vs legacy |
| `make test-engage-route-parity` | **PASS** | `scripts/engage/check-route-parity.py` |
| `make test-engage-tool-matrix` | **PASS** (best-effort) | 1/18 exercised locally (binaries missing on PATH); CI uses `enable-tools-on-path.sh` |
| `make test-engage-benchmark` | **SKIP** | engage-api not running on :8890 |
| `make test-engage-events-pipeline` | **PASS** | Phase 24: Neo4j poll `EngageToolRun`, health fail-fast + service logs; NATS + `graph-ingest` profile |
| `make test-engage-veil-stack-ci` | **PASS** | Phase 24: veil category search uses Neo4j 5–safe predicates in `knowledge/connector/query/service.go`; veil-stack-ci polls `nodes[]`, `SMOKE_VEIL_API_WAIT_SEC`/diagnostics; CI `engage-veil-stack` |

## MCP ↔ catalog ↔ runner

CSV: [engage-mcp-runner-triangle.csv](engage-mcp-runner-triangle.csv) (regenerate: `python3 scripts/engage/audit-mcp-runner-triangle.py`).

| Metric | Value |
|--------|-------|
| Enabled in `tools.live.yaml` | **158** rows (**104** subprocess + **54** bridge; 0 orphan — see [engage-tools-na-matrix.md](engage-tools-na-matrix.md)) |
| Runnable in runner image | Tier-1 subset when `engage-runner` image present (`list-runner-binaries.sh`); full port **P9i/j** |
| Tool matrix strict | `ENGAGE_TOOL_MATRIX_STRICT=1` via `make test-engage-runner-profile` (≥30 tools in runner container) |

## P2 HTTP backlog (closed in audit)

| Legacy route | Engage action |
|--------------|---------------|
| `POST /api/vuln-intel/attack-chains` | Alias → `discover-attack-chains` |
| `POST /api/vuln-intel/threat-feeds` | Wrapper → `cve-monitor` |
| `POST /api/vuln-intel/zero-day-research` | Heuristic stub (CVE lookup / discover chains) |
| `GET/POST /api/error-handling/*` | Read-only diagnostics API (`router_error_handling.go`) |

## Audit closure checklist

- [x] Automated gates green or documented (see table above)
- [x] `make test-engage-route-parity` — 0 unexplained missing
- [x] [engage-legacy-parity.md](engage-legacy-parity.md) route matrix + [engage-route-parity.csv](engage-route-parity.csv)
- [x] Master plan + greenfield synced (Phase 16–23; gap matrix updated)
- [x] P0: tier-1 live tools + `make test-engage-tool-matrix-strict` in compose smoke (superseded by **P9f/h/i** full-port KPIs — do not read as 158/158 execution done)
- [x] P2: vuln-intel aliases + `/api/error-handling/*` diagnostics

## Remaining backlog (post-audit)

| Priority | Item | Owner |
|----------|------|-------|
| P1 | CTF/BB golden JSON vs Python fixtures | Phase 24+ |
| P1 | `make test-engage-events-pipeline` flake (Neo4j count / cypher-shell parsing) | ops |
| Future | Findings FP/dedup labeled dataset | backlog |
| Future | README 24× speed as CI KPI | not in scope |

## Migration sign-off (Phase 30, R148–R150)

**HexStrike MCP + Flask `:8888` → veil-engage (docs-only closure).**

| Deliverable | Status |
|-------------|--------|
| R148 — Operator runbook: disable legacy MCP/`:8888`, configure **veil-engage** + env / Cursor ([mcp-agents.md](mcp-agents.md#migration-runbook-hexstrike-flask-8888--veil-engage)) | **Documented** |
| R149 — When `.external/` optional vs extract-only parity use ([external-hexstrike.md](external-hexstrike.md#when-external-is-optional)) | **Documented** |
| R150 — Formal sign-off in this section | **[x]** — 2026-05-16 |

**Signed-off criterion:** Teams can operate **without** Flask **`:8888`** per runbook; tool execution MCP is **`veil-engage`** only. This sign-off is **architecture + catalog/route parity**, not claim that every catalog tool subprocess-runs in the default runner image. Execution completeness is tracked separately: `make test-engage-executable-matrix` (**P9f**) per [engage_tools_full_coverage.plan.md](../.cursor/plans/engage_tools_full_coverage.plan.md).

## Client-native execution sign-off (Wave 5, P22)

**Date:** 2026-05-20

| Criterion | Status | Notes |
|-----------|--------|-------|
| Supported engage execution mode | **client-native only** | `ENGAGE_EXECUTION_PROFILE=client-native` is the default operational path (host PATH on MCP execution host). |
| Docker runner mode | **legacy lab/CI overlay** | `docker-exec` remains in `deploy/engage/compose.runner.yml` for isolated toolbox tests only; removed from default compose happy-path. |
| MCP operator contract | **Updated** | [mcp-agents.md](../agents/mcp-agents.md) documents host-local binary requirement and separates graph/read MCP from engage/exec MCP. |
| Master plan closure | **Wave 5 complete** | [engage_mcp_client_native_execution_master.plan.md](../.cursor/plans/engage_mcp_client_native_execution_master.plan.md) tracks P21–P22 completion. |

## Related docs

- [engage-legacy-parity.md](engage-legacy-parity.md) — living checklist + route matrix summary
- [engage-tools.md](engage-tools.md) — runner matrix
- [external-hexstrike.md](../external/external-hexstrike.md) — reference boundary
