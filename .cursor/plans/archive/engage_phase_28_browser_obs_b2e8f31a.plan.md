---
name: Engage Phase 28 Browser Obs
overview: "Phase 28 (R139–R142): browser forms/security parity, resource-usage API, benchmark CI artifact."
todos:
  - id: p28-r139-browser
    content: "R139: browser forms + security_analysis in sidecar/service"
    status: pending
  - id: p28-r140-resource
    content: "R140: GET /api/process/resource-usage via router_process.go"
    status: pending
  - id: p28-r141-benchmark
    content: "R141: benchmark JSON artifact + engage.yml upload"
    status: pending
  - id: p28-r142-smoke
    content: "R142: compose/benchmark smoke hook (SKIP ok)"
    status: pending
isProject: false
---

# Phase 28 — Browser & observability (R139–R142)

**Ветка:** `engage/phase-28-browser-obs`  
**Конфликт с Phase 29:** только **новый** `router_process.go` + 1 вызов `registerProcessRoutes` в `router.go`; не делать mass-split router.

## R139 — Browser

- Extend [browser/service.go](engage/serve/internal/usecase/browser/service.go) + [cmd/browser-agent/index.mjs](engage/serve/cmd/browser-agent/index.mjs): `forms[]`, `security_analysis` subset.
- Tests in `browser/service_test.go`.
- `make test-engage-browser` green when Docker profile available (document SKIP).

## R140 — Resource usage

- New [router_process.go](engage/serve/internal/transport/httpserver/router_process.go): `GET /api/process/resource-usage` delegating to process manager / telemetry.
- Minimal edit `router.go` (register call only).

## R141–R142 — Benchmark

- [scripts/benchmark/engage-hexstrike-parity.sh](scripts/benchmark/engage-hexstrike-parity.sh): write `scripts/benchmark/results/latest.json`.
- CI: upload artifact in engage.yml (no 24× gate).
- Optional: `make test-engage-compose` calls benchmark with SKIP if API down.

## DoD

- [ ] `make test-engage` green
- [ ] browser unit tests green
- [ ] benchmark script produces JSON when `ENGAGE_API_URL` set

## Verify

```bash
make test-engage
cd engage/serve && go test ./internal/usecase/browser/... -count=1
```
