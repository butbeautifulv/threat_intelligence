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
    content: "discovery/harvest/internal/scrapepub base; migrate ti→vuln→ds"
    status: completed
  - id: p6d-makefile-smoke-lib
    content: "scripts/test/lib/smoke.sh + Makefile test-pkg-shared"
    status: completed
  - id: p6e-deploy-stack-presets
    content: "deploy/stacks presets SSOT for compose overlays"
    status: completed
  - id: p6f-engage-package-split
    content: "Split intel_bridge, analyze.go, findings/parse.go"
    status: completed
  - id: p6g-natsjet-envelope-pub
    content: "pkg/natsjet envelope publish helper; thin scrape/pipeline connectors"
    status: completed
isProject: false
---

# Veil Platform P6 — Maximum safe refactor

## Constraints (unchanged)

- No cross-imports: `discovery/`, `pipeline/`, `graph/`, `engage/`.
- Shared wire: `pkg/harvest`, `pkg/commit`, `pkg/natsjet`, `pkg/auth`, `pkg/engage/*`.
- Engage → graph: HTTP only.

## Completed (this branch)

| Item | Change |
|------|--------|
| **pkg/engage/events** | `AuditEvent`, `FindingEvent`, `Publisher` via `natsjet` |
| **pkg/auth/httpmiddleware** | Single JWT+RBAC wrapper; engage/graph 3-line adapters |
| **Removed** | `engage/serve/internal/events`, `events_bridge` |
| **Makefile** | `test-pkg-shared` phony |

## Completed (P6c)

| Item | Change |
|------|--------|
| **discovery/harvest/internal/scrapepub** | `RawPublisher`, `Base`, `NewRaw`; factory type alias |
| **Migrated** | ti, vuln, ds, lola per-source scrapepub embed shared base |

## Completed (P6d–g, merged `505806a`)

| Item | Change |
|------|--------|
| **P6d** | `smoke_wait_http` label arg; 4 platform smokes source `lib/smoke.sh` |
| **P6e** | `deploy/stacks/{minimal,full-loop,secure*,pentest-prod}.yml` + README |
| **P6f** | Split `intel_bridge_*`, `findings/parse_*`, `intelligence/{attack_chain,select_tools}` |
| **P6g** | `pkg/natsjet/publish.go`; thin scrape/pipeline connector publish |

## Next batches

- Engage Phase 29 / platform v4 CI — see master engage & v4 plans

## Verification per batch

```bash
make test-pkg-shared
make test-engage
make test-pipeline
make test-scrape   # after P6c
make test-platform-p0
```
