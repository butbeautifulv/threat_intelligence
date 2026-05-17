# Veil domain contour (pkg SOT)

Shared threat-intelligence types and rules live under `pkg/`. Runtime layers (`discovery/`, `pipeline/`, `graph/`, `engage/`) keep source-specific adapters only.

## Package map

| Package | Role | Imported by |
|---------|------|-------------|
| `pkg/ti/domain` | IOC, Actor, Campaign, Cluster, Report | scrape TI, pipeline NED, graph ingest TI |
| `pkg/vuln/domain` | Vulnerability, CVSS, CPE, ExploitRef | scrape vuln, pipeline NED, graph ingest vuln |
| `pkg/lola/domain` | Artifact, Command, Detection, … | scrape lola, graph ingest lola |
| `pkg/ds/domain` | Resource (Caldera, Sigma, …) | scrape ds, pipeline NED, graph ingest ds |
| `pkg/sbom/domain` | AdvisoryRef (OSV/GHSA) | scrape sbom, pipeline NED, graph ingest sbom |
| `pkg/nuclei/domain` | Template | scrape nuclei, pipeline NED, graph ingest nuclei |
| `pkg/coderules/domain` | RuleFile (Semgrep, CodeQL) | scrape coderules, pipeline NED, graph ingest coderules |
| `pkg/ti/validate` | Pure validation (type, empty fields) | `pkg/ti/normalize` |
| `pkg/ti/normalize` | NED normalization (IOC canonical form) | **pipeline only** — not graph ingest |
| `pkg/ti/ids` | Stable actor/report/IOC ids for dedup | pipeline NED (`normalize` re-exports) |
| `pkg/harvest` | Scrape → NATS envelope | scrape, pipeline |
| `pkg/commit` | Pipeline → graph ingest envelope | pipeline, graph |
| `pkg/engage/contract` | Engage HTTP/MCP DTOs | engage serve |
| `pkg/engage/domain/report` | Finding, Severity | engage serve |
| `pkg/engage/domain/job` | Job, Status | engage serve |
| `pkg/engage/domain/tool` | Tool spec, catalog metadata | engage serve |
| `pkg/engage/events` | Finding events wire | engage, pipeline, graph ingest |

## Layer adapters (not in pkg)

| Layer | Path pattern | Responsibility |
|-------|--------------|----------------|
| Scrape | `discovery/harvest/internal/sources/<src>/` | Fetch, parse → `harvest.Envelope` (uses `pkg/*/domain`) |
| Pipeline | `pipeline/ned/internal/sources/<src>/` | Transform → normalize → `commit.Envelope` |
| Graph ingest | `graph/ingest/internal/sources/<src>/` | Apply commit → Neo4j (uses `pkg/*/domain`) |
| Engage | `engage/serve/internal/domain/target` | Target allowlist / guard (report, job, tool in `pkg/engage/domain`) |

## TI flow

```text
scrape (raw IOC in harvest payload)
  → pipeline NED (pkg/ti/normalize + pkg/ti/ids)
  → commit envelope (normalized payload)
  → graph ingest (MERGE, no re-normalize)
```

## Deprecations

- `pipeline/pkg/ti/normalize` — thin forwarder to `pkg/ti/normalize`; new code should import `pkg/ti/normalize` directly.
