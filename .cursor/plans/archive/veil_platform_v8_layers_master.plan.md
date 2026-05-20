---
name: Veil Platform v8 Layers Master
overview: "Done (main 1f40c77): discovery, knowledge, pipeline, engage, pkg/report, pkg/decision, pkg/exec, pkg/api, pkg/mcp, engage slim, discovery browser."
todos:
  - id: v8a-domain-complete
    content: "P8a: Завершить pkg/domain — engage target, knowledge read models, zero dup structs"
    status: completed
  - id: v8b-pkg-report
    content: "P8b: pkg/report — HTML/PDF/executive; engage thin adapter"
    status: completed
  - id: v8c-pkg-decision
    content: "P8c: pkg/decision — intel engine, attack chain, tool selection (from engage)"
    status: completed
  - id: v8d-api-mcp-facade
    content: "P8d: pkg/api + pkg/mcp — shared transport; graph + optional engage mount"
    status: completed
  - id: v8e-pkg-exec
    content: "P8e: pkg/exec — Sandbox/Executor/audit; engage migrate; scrape fetcher profile optional"
    status: completed
  - id: v8f-engage-slim
    content: "P8f: Engage slim — pentest tools, workflows, guard; drop report/decision/browser"
    status: completed
  - id: v8g-discovery-browser
    content: "P8g: Discovery — browser crawl worker; factory stays orchestration"
    status: completed
  - id: v8h-rename-scrape-discovery
    content: "P8h: Rename module scrape/ → discovery/ (paths, deploy, Makefile, CI)"
    status: completed
  - id: v8i-rename-graph-knowledge
    content: "P8i: Rename module graph/ → knowledge/ (paths, deploy, Makefile, CI)"
    status: completed
isProject: false
---

# Veil Platform v8 — logical layers master plan

**Prerequisite (done):** P6 refactor, P7 tests + `pkg/*/domain`, HexStrike sign-off, catalog live merge fix (`634e067`).

**Architecture doc:** [docs/platform-architecture.md](../../docs/platform-architecture.md)

## Target layer map

| Layer | Path today | Path target (v8) | Notes |
|-------|------------|------------------|--------|
| **Discovery** | `discovery/` | **`discovery/`** | P8h rename; `harvest` wire unchanged |
| **Pipeline** | `pipeline/` | `pipeline/` | name kept (NED role is clear) |
| **Knowledge** | `knowledge/` | **`knowledge/`** | P8i rename; veil-api / veil-mcp brands may stay |
| **Engage** | `engage/` | `engage/` | offensive tools layer name kept |
| **Report** | engage `report/` | **`pkg/report`** | P8b |
| **API/MCP** | layer transports | **`pkg/api`**, **`pkg/mcp`** | P8d |

**Naming rule:** documentation and repo top-level dirs use **Discovery** / **Knowledge**; until P8h/P8i merge, legacy paths `discovery/` and `knowledge/` remain valid in code.

## Constraints

- No cross-import between runtime modules (today: `scrape`, `pipeline`, `graph`, `engage`; after rename: `discovery`, `pipeline`, `knowledge`, `engage`).
- **Wire-stable (P8h/P8i):** NATS subjects (`scrape.>`, `ingest.>`), `pkg/harvest` / `pkg/commit` JSON field names (`source: ti|vuln|…`), graph pack IDs — rename dirs only unless a dedicated schema version bump.
- Engage → knowledge read API: HTTP veil-api only (URL/env may keep `VEIL_API` / `ENGAGE_VEIL_API_URL` during transition).
- **Do not** rename scrape `factory` to `runner`; share **`pkg/exec`** primitives only.

---

## P8a — Complete pkg/domain

**Branch:** `platform/p8a-domain-complete`

- [ ] `engage/serve/internal/domain/target` → `pkg/engage/domain/target` (if shared)
- [ ] Graph serve read DTOs: document layer-local vs `pkg/`
- [ ] `grep` — no duplicate `type IOC struct` outside `pkg/ti/domain`
- [ ] `make test-platform-p7`, `make test-pkg-domain`

---

## P8b — pkg/report

**Branch:** `platform/p8b-pkg-report`

- [ ] Extract `engage/serve/internal/usecase/report` → `pkg/report/` (templates, render ports)
- [ ] Engage HTTP handlers call `pkg/report` adapters
- [ ] Optional: knowledge/pipeline consumers later (same package, no layer import)
- [ ] Tests: golden HTML fragment tests in `pkg/report`

**DoD:** `make test-engage`; no Neo4j in `pkg/report`.

---

## P8c — pkg/decision

**Branch:** `platform/p8c-pkg-decision`

- [ ] Move pure logic from `engage/.../intelligence/` (`select_tools`, `attack_chain`, effectiveness tables)
- [ ] `IntelligenceProvider` port stays in engage; implementation uses `pkg/decision`
- [ ] `make test-engage-decision-parity` green

**DoD:** engage intelligence package <300 LOC (wiring only).

---

## P8d — pkg/api + pkg/mcp

**Branch:** `platform/p8d-api-mcp-facade`

- [ ] Shared route registration helpers, auth wrapper (`pkg/auth` already)
- [ ] MCP tool/list/call framing shared; engage vs graph plugin mounts
- [ ] Single doc: which MCP exposes graph-only vs engage-only vs combined gateway (optional future `veil-gateway` binary)

**DoD:** `make test-graph-serve`, `make test-engage`, route parity unchanged.

---

## P8e — pkg/exec (runner/sandbox plane)

**Branch:** `platform/p8e-pkg-exec`

- [ ] Extract from `engage/serve/internal/runner`: `Sandbox`, `Executor`, `ProcessTracker`, allowlist env
- [ ] Engage `runner` package becomes thin adapter
- [ ] **Scrape spike:** `discovery-fetcher` compose profile — one source (e.g. git clone or headless) using same Sandbox
- [ ] Document when scrape should **not** use exec (plain HTTP feeds — keep `feeds.Client`)

**DoD:** engage runner tests pass; scrape spike behind build tag `discoveryexec`.

---

## P8f — Engage slim

**Branch:** `platform/p8f-engage-slim`

- [ ] Remove moved report/decision code from engage
- [ ] Catalog + `tools.Runner` + workflows + MCP intel bridge wiring
- [ ] `make test-engage-hardening`, `make test-engage-parity`

---

## P8g — Discovery browser

**Branch:** `platform/p8g-discovery-browser`

- [ ] Move browser-agent under `discovery/` or `cmd/discovery-browser/`
- [ ] Publish harvest-compatible events (new kind or reuse existing)
- [ ] Engage deprecates direct browser service (HTTP proxy only if needed)

---

## P8h — Rename `discovery/` → `discovery/`

**Branch:** `platform/p8h-rename-discovery`  
**Depends on:** P8a recommended (fewer import churn); may run parallel to P8i if touch-disjoint.

| Area | Action |
|------|--------|
| Repo | `git mv scrape discovery`; update `discovery/go.work`, module paths `github.com/butbeautifulv/veil/discovery/...` |
| Deploy | `deploy/discovery/` → `deploy/discovery/`; compose service names (`discovery_worker` alias or replace `scrape_worker`) |
| Makefile | `test-scrape` → `test-discovery` (+ temporary alias `test-scrape` → prints deprecate) |
| CI | `.github/workflows/*` paths; `platform-p7` scrape slice → `test-discovery-p7c` |
| Docs | README, coding-style, runtime docs; **logical name Discovery everywhere** |
| Scripts | `scripts/test/smoke-scrape-*` → `smoke-discovery-*` (symlink or mv) |

**Keep unchanged in P8h:** `pkg/harvest` package name; NATS subject `scrape.>`; harvest envelope `source` enum strings; binary name `scrape_worker` optional alias one release.

**DoD:** `make test-discovery`, `make test-platform-p7`, `discovery/harvest` build; no remaining imports of `veil/scrape` except changelog.

---

## P8i — Rename `knowledge/` → `knowledge/`

**Branch:** `platform/p8i-rename-knowledge`  
**Depends on:** ingest tests green; **serialize merge with P8h** on `main` (both touch Makefile/root docs) or one combined `platform/p8hi-layer-rename` branch.

| Area | Action |
|------|--------|
| Repo | `git mv graph knowledge`; module `github.com/butbeautifulv/veil/knowledge/...` |
| Deploy | `deploy/knowledge/` → `deploy/knowledge/`; Neo4j compose paths |
| Binaries / brands | `veil-api`, `veil-mcp`, `ingest_worker` — **keep user-facing names**; only repo path `knowledge/` |
| Makefile | `test-graph` → `test-knowledge`; `test-graph-serve` → `test-knowledge-serve`; aliases one release |
| CI | engage.yml path filters `knowledge/` → `knowledge/` |
| Docs | `docs/threatintel-runtime.md` → split or rename `knowledge-runtime.md`; graph-pack.md may keep “graph pack” as artifact name |
| Versions | `GRAPH_PACK_VERSION` env key unchanged in P8i (rename env is P8i-follow-up optional) |

**Keep unchanged in P8i:** NATS `ingest.>`; `pkg/commit`; Neo4j labels; HTTP routes `/v1/*`; MCP server name `veil-mcp` in agent configs.

**DoD:** `make test-knowledge`, `make test-knowledge-serve`, `make check-graph-version` (script name alias OK); engage veilgraph client still reaches veil-api.

---

## Rename merge order (orchestrator)

```text
main ─┬─ P8h (discovery/) ──┐
      └─ P8i (knowledge/) ─┴─► merge rename wave ─► P8b–g on renamed paths
```

1. Finish **P8h + P8i** (or single branch) before wide P8b–g refactors to avoid double rebase.
2. One commit per rename layer; run full platform tests after each.
3. Update [platform-architecture.md](../../docs/platform-architecture.md) “Path today” column to only target names.

---

## Parallelism

| Parallel safe | Serial after |
|---------------|--------------|
| P8a + P8b + P8c (different pkg dirs) | P8f after b,c |
| P8e after P8a | P8g after P8e spike |
| P8d after P8a | |
| **P8h ∥ P8i** (disjoint trees) | **merge both before P8b–g** or rebase P8b–g onto renamed `main` |
| P8b–g | prefer **after P8h+P8i** on `main` |

## Verification (each merge)

```bash
make test-platform-p7
make test-pkg-shared
make test-engage
make test-discovery    # after P8h (alias: test-scrape)
make test-pipeline
make test-knowledge    # after P8i (alias: test-graph)
make check-graph-version   # if ingest touched
```

---

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P8a | `platform/p8a-domain-complete` | done — `7ac8da4` |
| P8b | `platform/p8b-pkg-report` | done — `9f3513f` |
| P8c | `platform/p8c-pkg-decision` | done — `5d7bc03` |
| P8d | `platform/p8d-api-mcp-facade` | done — `7fb348e` |
| P8e | `platform/p8e-pkg-exec` | done — `b40f266` |
| P8f | `platform/p8f-engage-slim` | done — `2650549` |
| P8g | `platform/p8g-discovery-browser` | done — `a7bf1cd` |
| P8h | `platform/p8h-rename-discovery` | done — merge `24af6ad` |
| P8i | `platform/p8i-rename-knowledge` | done — merge `24af6ad` |
