---
name: Veil Platform v8 Layers Master
overview: "Целевая архитектура: discovery, knowledge, pipeline, engage, shared report, unified API/MCP; завершение pkg/domain; общий execution plane (runner/sandbox) без слияния с scrape factory."
todos:
  - id: v8a-domain-complete
    content: "P8a: Завершить pkg/domain — engage target, graph read models, zero dup structs"
    status: pending
  - id: v8b-pkg-report
    content: "P8b: pkg/report — HTML/PDF/executive; engage thin adapter"
    status: pending
  - id: v8c-pkg-decision
    content: "P8c: pkg/decision — intel engine, attack chain, tool selection (from engage)"
    status: pending
  - id: v8d-api-mcp-facade
    content: "P8d: pkg/api + pkg/mcp — shared transport; graph + optional engage mount"
    status: pending
  - id: v8e-pkg-exec
    content: "P8e: pkg/exec — Sandbox/Executor/audit; engage migrate; scrape fetcher profile optional"
    status: pending
  - id: v8f-engage-slim
    content: "P8f: Engage slim — pentest tools, workflows, guard; drop report/decision/browser"
    status: pending
  - id: v8g-discovery-browser
    content: "P8g: Discovery — browser crawl worker; factory stays orchestration"
    status: pending
isProject: false
---

# Veil Platform v8 — logical layers master plan

**Prerequisite (done):** P6 refactor, P7 tests + `pkg/*/domain`, HexStrike sign-off, catalog live merge fix (`634e067`).

**Architecture doc:** [docs/platform-architecture.md](../../docs/platform-architecture.md)

## Target layer map

| Layer | Module(s) today | v8 outcome |
|-------|-----------------|------------|
| Discovery | `scrape/` | `scrape/` + optional sandboxed fetcher; browser out of engage |
| Pipeline | `pipeline/` | unchanged role; more transforms use `pkg/*` |
| Knowledge | `graph/` + engage intel | `graph/` + **`pkg/decision`** |
| Engage | `engage/` | catalog, runner, workflows, guard only |
| Report | engage `report/` | **`pkg/report`** |
| API/MCP | graph + engage servers | **`pkg/api`**, **`pkg/mcp`** |

## Constraints

- No cross-import `scrape` / `pipeline` / `graph` / `engage`.
- NATS + `pkg/harvest`, `pkg/commit`, `pkg/engage/events` unchanged unless versioned.
- Engage → graph: HTTP veil-api only.
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
- [ ] Optional: graph/pipeline consumers later (same package, no layer import)
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

- [ ] Move browser-agent under `scrape/` or `cmd/discovery-browser/`
- [ ] Publish harvest-compatible events (new kind or reuse existing)
- [ ] Engage deprecates direct browser service (HTTP proxy only if needed)

---

## Parallelism

| Parallel safe | Serial after |
|---------------|--------------|
| P8a + P8b + P8c (different pkg dirs) | P8f after b,c |
| P8e after P8a | P8g after P8e spike |
| P8d after P8a | |

## Verification (each merge)

```bash
make test-platform-p7
make test-pkg-shared
make test-engage
make test-scrape
make test-pipeline
make test-graph
make check-graph-version   # if ingest touched
```

---

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P8a | `platform/p8a-domain-complete` | pending |
| P8b | `platform/p8b-pkg-report` | pending |
| P8c | `platform/p8c-pkg-decision` | pending |
| P8d | `platform/p8d-api-mcp-facade` | pending |
| P8e | `platform/p8e-pkg-exec` | pending |
| P8f | `platform/p8f-engage-slim` | pending |
| P8g | `platform/p8g-discovery-browser` | pending |
