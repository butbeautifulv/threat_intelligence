---
name: Veil Platform P6 Refactor
overview: "Максимальный безопасный рефакторинг платформы без нарушения четырёх контекстов: pkg/* SOT, NATS между слоями, HTTP engage→graph."
todos:
  - id: p6a-pkg-engage-events
    content: "pkg/engage/events wire + natsjet publisher; pipeline consumer"
    status: completed
  - id: p6b-pkg-auth-http
    content: "pkg/auth/httpmiddleware; thin engage/graph wrappers"
    status: completed
  - id: p6c-scrape-scrapepub
    content: "scrape/harvest/internal/scrapepub base; migrate ti→vuln→ds"
    status: pending
  - id: p6d-makefile-smoke-lib
    content: "scripts/test/lib/smoke.sh + Makefile test-pkg-shared"
    status: in_progress
  - id: p6e-deploy-stack-presets
    content: "deploy/stacks presets SSOT for compose overlays"
    status: pending
  - id: p6f-engage-package-split
    content: "Split intel_bridge, analyze.go, findings/parse.go"
    status: pending
  - id: p6g-natsjet-envelope-pub
    content: "pkg/natsjet envelope publish helper; thin scrape/pipeline connectors"
    status: pending
isProject: false
---

# Veil Platform P6 — Maximum safe refactor

## Constraints (unchanged)

- No cross-imports: `scrape/`, `pipeline/`, `graph/`, `engage/`.
- Shared wire: `pkg/harvest`, `pkg/commit`, `pkg/natsjet`, `pkg/auth`, `pkg/engage/*`.
- Engage → graph: HTTP only.

## Completed (this branch)

| Item | Change |
|------|--------|
| **pkg/engage/events** | `AuditEvent`, `FindingEvent`, `Publisher` via `natsjet` |
| **pkg/auth/httpmiddleware** | Single JWT+RBAC wrapper; engage/graph 3-line adapters |
| **Removed** | `engage/serve/internal/events`, `events_bridge` |
| **Makefile** | `test-pkg-shared` phony |

## Next batches (priority)

1. **P6c** — `scrape/harvest/internal/scrapepub` (DRY per-source publishers)
2. **P6d** — `scripts/test/lib/smoke.sh` (wait_http, skip_no_docker)
3. **P6e** — `deploy/stacks/*.yml` compose preset chains
4. **P6f** — engage large file splits (in-layer only)
5. **P6g** — `pkg/natsjet` typed publish facade for scrape/pipeline connectors

## Verification per batch

```bash
make test-pkg-shared
make test-engage
make test-pipeline
make test-scrape   # after P6c
make test-platform-p0
```
