---
name: Engage Phase 5 R20
overview: "Слайс R20 — smart-scan: AnalyzeTarget → SelectTools[:max_tools] → parallel execution (sync goroutines or async jobs)."
todos:
  - id: r20-impl
    content: "workflow.SmartScan + router + tests"
    status: pending
isProject: false
---

# Engage Phase 5 — R20 (Smart scan)

## Flow

`POST /api/intelligence/smart-scan` body: `target`, `objective`, `max_tools` (default 5), `async` (default false).

1. `AnalyzeTarget` → `SelectTools` capped by `max_tools` and `objective`.
2. `async=false`: bounded parallel `Tools.Run` (max 5 workers).
3. `async=true`: `Jobs.Enqueue` per tool, return job IDs.

## Files

- [`workflow/smartscan.go`](engage/serve/internal/usecase/workflow/smartscan.go)
- [`router.go`](engage/serve/internal/transport/httpserver/router.go) — wire `SmartScan`
- [`router_test.go`](engage/serve/internal/transport/httpserver/router_test.go)
