# Veil domain contour (pkg SOT)

Shared threat-intelligence types and rules live under `pkg/`. Runtime layers (`scrape/`, `pipeline/`, `graph/`, `engage/`) keep source-specific adapters only.

## Package map

| Package | Role | Imported by |
|---------|------|-------------|
| `pkg/ti/domain` | IOC, Actor, Campaign, Cluster, Report | scrape (aliases), pipeline NED, graph ingest |
| `pkg/ti/validate` | Pure validation (type, empty fields) | `pkg/ti/normalize` |
| `pkg/ti/normalize` | NED normalization (IOC canonical form) | **pipeline only** — not graph ingest |
| `pkg/ti/ids` | Stable actor/report/IOC ids for dedup | pipeline NED (`normalize` re-exports) |
| `pkg/harvest` | Scrape → NATS envelope | scrape, pipeline |
| `pkg/commit` | Pipeline → graph ingest envelope | pipeline, graph |
| `pkg/engage/contract` | Engage HTTP/MCP DTOs | engage serve |
| `pkg/engage/events` | Finding events wire | engage, pipeline, graph ingest |

## Layer adapters (not in pkg)

| Layer | Path pattern | Responsibility |
|-------|--------------|----------------|
| Scrape | `scrape/harvest/internal/sources/<src>/` | Fetch, parse feeds → `harvest.Envelope` |
| Pipeline | `pipeline/ned/internal/sources/<src>/` | Transform → normalize TI → `commit.Envelope` |
| Graph ingest | `graph/ingest/internal/sources/<src>/` | Apply commit → Neo4j (expects NED-normalized TI) |
| Engage | `engage/serve/internal/domain/` | Tool specs, jobs, findings (not wire DTOs) |

## TI flow

```text
scrape (raw IOC in harvest payload)
  → pipeline NED (pkg/ti/normalize + pkg/ti/ids)
  → commit envelope (normalized payload)
  → graph ingest (MERGE, no re-normalize)
```

## Deprecations

- `pipeline/pkg/ti/normalize` — thin forwarder to `pkg/ti/normalize`; new code should import `pkg/ti/normalize` directly.
