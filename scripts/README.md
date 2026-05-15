# Scripts

Host-side helpers for **Neo4j export**, **graph pack** build/import, **stack smoke**, and **housekeeping**. Runtime layout: [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md).

These scripts are **not** the pipeline NED runtime ([pipeline/ned](../pipeline/ned/)); they orchestrate compose, verify graph state, or repair Neo4j after ingest.

| Script | Layer | Purpose |
|--------|-------|---------|
| [compose-up-full.sh](compose-up-full.sh) | ops | Full stack: scrape + pipeline + graph (`deploy/*/compose.yml`) |
| [graph-pack-run-v032.sh](graph-pack-run-v032.sh) | ops | Fast-rich pack profile: all 7 sources, `NVD_MAX_PAGES=1`, `LOFTS_SKIP_ON_ERROR` |
| [smoke_scrape_e2e.sh](smoke_scrape_e2e.sh) | ops | E2E smoke: scrape → pipeline → ingest → Neo4j (default `SCRAPE_SOURCES=ti,sbom`) |
| [verify-nvd-enrichment.sh](verify-nvd-enrichment.sh) | graph QA | Cypher counts for `HAS_CWE` / `AFFECTS` / `CPE` after NVD ingest |
| [export-graph-cypher.sh](export-graph-cypher.sh) | ops | Dump Cypher from running Neo4j (needs `NEO4J_apoc_export_file_enabled`) |
| [build-graph-pack.sh](build-graph-pack.sh) | ops | Build versioned ZIP + `manifest.json` + checksum |
| [import-graph-pack.sh](import-graph-pack.sh) | ops | Import a pack ZIP into Neo4j |
| [graph-dedup-cleanup.sh](graph-dedup-cleanup.sh) | graph housekeeping | Post-scrape dedup and optional stale IOC cleanup |

Graph pack and export scripts use the same compose files as smoke (override with `COMPOSE_FILES`):

```bash
COMPOSE_FILES="-f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml"
```

Quick smoke (minimal scrape, no heavy downloads):

```bash
SCRAPE_SOURCES=ti SMOKE_CLEAN_VOLUMES=0 ./scripts/smoke_scrape_e2e.sh --up
```

---

## `graph-dedup-cleanup.sh`

Neo4j housekeeping after high-volume scrapes. **Not** pipeline wire-path dedup (that is `commit` idempotency keys in [pipeline/ned](../pipeline/ned/)).

- **Duplicate `HAS_ADVISORY`:** removes parallel edges between the same `Vulnerability` and `SecurityAdvisory`.
- **Stale isolated IOCs:** counts `IOC` nodes with **no relationships** whose `lastSeen` or `updatedAt` is older than a cutoff (default **90 days**). Optional **destructive** delete is **off by default**.

### Usage

```bash
./scripts/graph-dedup-cleanup.sh --dry-run
./scripts/graph-dedup-cleanup.sh
GRAPH_DELETE_STALE_ISOLATED_IOCS=1 GRAPH_IOC_STALE_DAYS=120 ./scripts/graph-dedup-cleanup.sh
```

Requires `cypher-shell` on `PATH`, or run via `docker compose` against the same project as the running stack.

### Environment

| Variable | Default | Meaning |
|----------|---------|--------|
| `NEO4J_URI` | `neo4j://localhost:7687` | Bolt/Neo4j URI |
| `NEO4J_USER` / `NEO4J_PASS` / `NEO4J_DB` | `neo4j` / `neo4jpassword` / `neo4j` | Auth |
| `GRAPH_IOC_STALE_DAYS` | `90` | Age threshold for stale isolated IOCs |
| `GRAPH_DELETE_STALE_ISOLATED_IOCS` | `0` | Set to `1` to delete stale isolated IOCs when not using `--dry-run` |
