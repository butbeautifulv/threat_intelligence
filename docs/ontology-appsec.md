# AppSec graph ontology (threat_intelligence)

This document summarizes **normalized node labels**, **relationships** introduced by scrape sources, and the **roadmap** for DAST / SOC classes. It complements live category metadata from the graph API layer.

## Four data classes (coverage map)

| Class | Meaning | Primary labels | Scrapers |
|--------|---------|----------------|----------|
| Vulnerability intelligence | What is vulnerable and how it is exploited | `Vulnerability`, `CWE`, `CPE`, `Exploit` | `vuln`, `ti` (KEV context), `sbom` (package/advisory links) |
| SAST / code patterns | Dangerous code patterns and static rules | `CWE`, `SemgrepRule`, `CodeQLRule` | `coderules` |
| DAST / runtime checks | Network/runtime checks and template-based verification | `NucleiTemplate` | `nuclei` |
| SBOM / dependency intelligence | Packages and published advisories | `Package`, `SecurityAdvisory` | `sbom` |

## Core relationships (high value)

- `(:Vulnerability)-[:HAS_CWE]->(:CWE)` — from NVD (`vuln`).
- `(:Vulnerability)-[:AFFECTS_PACKAGE]->(:Package)` — from OSV (`sbom`).
- `(:Vulnerability)-[:HAS_ADVISORY]->(:SecurityAdvisory)` — from GHSA (`sbom`).
- `(:SecurityAdvisory)-[:AFFECTS_PACKAGE]->(:Package)` — from GHSA (`sbom`).
- `(:SecurityAdvisory)-[:ADVISORY_MAPS_TO_CWE]->(:CWE)` — from GHSA (`sbom`).
- `(:SemgrepRule)-[:MAPS_TO_CWE]->(:CWE)` — when Semgrep rule `metadata.cwe` lists identifiers (`coderules`).
- `(:CodeQLRule)-[:MAPS_TO_CWE]->(:CWE)` — when `CWE-*` tokens appear in the query preamble (`coderules`).
- `(:NucleiTemplate)-[:RELATES_TO_CVE]->(:Vulnerability)` — when the template references a CVE present in the graph (`nuclei`).
- `(:NucleiTemplate)-[:MAPS_TO_CWE]->(:CWE)` — when `classification.cwe-id` is set (`nuclei`).

## Anti–data-swamp principles

- Hard **limits** via environment variables on every high-cardinality feed (`*_MAX_*`).
- **MERGE** on canonical keys (`id`, `cve`, `Package.id`, `SemgrepRule.id`, …).
- **NATS JetStream** path: scrape sources publish **`scrapev1`** to **`scrape.>`**; **`pipeline_worker`** normalizes to **`ingestv1`** on **`ingest.>`**; **`ingest_worker`** MERGEs into Neo4j (AppSec via `graph/storage/*`, domains via `graph/sources/*`). Subject matrix: [ingest-contract.md](ingest-contract.md).
- Optional **cleanup** scripts under [`scripts/`](../scripts/) (duplicate relationships, stale isolated IOCs) with `--dry-run` first.

## IOC freshness (TI)

IOC nodes store `firstSeen`, `lastSeen`, `sources`, and `updatedAt` (see [scrape/sources/ti](../scrape/sources/ti)). Feeds with fast-moving indicators (URLhaus, OpenPhish) should be aged or reaped using documented Cypher thresholds—not by implicit deletes in the write path.

## P3 roadmap (SOC-level rules, not implemented as scrapers yet)

Planned as **separate ontologies** after P1–P2 stabilize:

- **OWASP ModSecurity CRS** — WAF rule bundles; likely `WafRule` or `ModSecurityRule` label family.
- **Snort / Suricata** — network IDS rules; watch vendor licensing (e.g. Emerging Threats).
- **Zeek** — scripting/detections as `ZeekScript` or similar.

### Inclusion criteria (before adding a P3 scraper)

- **Stable public source** — durable URLs or versioned artifacts, predictable updates.
- **Clear stable key** — one natural primary key per normalized node (or deterministic hash of source + key).
- **Acceptable graph cardinality** — volume bounded by env limits and/or subset feeds; no “clone entire ruleset” by default without an explicit full-sync profile.
- **License compatible** with how you redistribute graph packs / exports (commercial feeds need explicit approval).

## SAST extras (roadmap only)

Future adapters (not in scope for the first AppSec graph release): Joern traversals, Sonar rules, PMD, ESLint security plugins, Bandit, gosec, Roslyn analyzers, Facebook Infer. Same inclusion criteria as P3; prefer separate labels or a thin `(:Rule {engine})` pattern once one adapter is proven.

## CWE hierarchy (optional future)

The MITRE CWE XML includes **parent/child** relationships between weaknesses. This repo currently enriches `(:CWE)` with catalog fields (`name`, `description`, `status`, …). Adding typed edges such as `(:CWE)-[:PARENT_OF]->(:CWE)` is a follow-up when read APIs need hierarchy navigation.

## Correlation layer (future)

A read-side or batch **enrichment engine** (outside the Neo4j write path) can materialize “attack path” subgraphs: CVE → package → advisory → CWE → Semgrep/CodeQL/Nuclei → detection content. This repo currently focuses on **ingest + categorical query APIs**; correlation jobs can be added later without changing ingest contracts.

## Related documentation

- [threatintel-runtime.md](threatintel-runtime.md) — Compose, API, NATS, **`ingest_worker`**
- [scrape/README.md](../scrape/README.md) — scrape sources and env vars
- [graph/ingest_worker/README.md](../graph/ingest_worker/README.md) — graph consumer
- [deploy.md](deploy.md) — worker scaling
- [coding-style.md](coding-style.md) — three-layer layout
