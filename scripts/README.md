# Scripts

Host-side helpers for **Neo4j export**, **graph pack** build/import, and **housekeeping**. Compose services (including **`ingest-worker`**) and **smoke / stack** procedures: [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md), especially *Scrape / pack-build profile* and *Smoke checklist*.

| Script / doc | Purpose |
|----------------|---------|
| [export-graph-cypher.sh](export-graph-cypher.sh) | Dump Cypher from a running Neo4j (used before `build-graph-pack.sh`) |
| [build-graph-pack.sh](build-graph-pack.sh) | Build versioned ZIP + `manifest.json` + checksum |
| [import-graph-pack.sh](import-graph-pack.sh) | Import a pack ZIP into Neo4j |
| **Below:** `graph-dedup-cleanup.sh` | Post-scrape dedup and optional stale IOC cleanup |

---

## `graph-dedup-cleanup.sh`

Neo4j housekeeping after high-volume scrapes.

- **Duplicate `HAS_ADVISORY`:** removes parallel edges between the same `Vulnerability` and `SecurityAdvisory`.
- **Stale isolated IOCs:** counts `IOC` nodes with **no relationships** whose `lastSeen` or `updatedAt` is older than a cutoff (default **90 days**). Optional **destructive** delete is **off by default**.

### Usage

```bash
# Preview only (no writes except read queries for counts)
./scripts/graph-dedup-cleanup.sh --dry-run

# Fix duplicate advisory edges
./scripts/graph-dedup-cleanup.sh

# Also remove stale degree-0 IOCs (use after reviewing counts)
GRAPH_DELETE_STALE_ISOLATED_IOCS=1 GRAPH_IOC_STALE_DAYS=120 ./scripts/graph-dedup-cleanup.sh
```

Requires `cypher-shell` on `PATH`, or run the printed `docker compose exec neo4j cypher-shell ...` variant.

### Environment

| Variable | Default | Meaning |
|----------|---------|--------|
| `NEO4J_URI` | `neo4j://localhost:7687` | Bolt/Neo4j URI |
| `NEO4J_USER` / `NEO4J_PASS` / `NEO4J_DB` | `neo4j` / `neo4jpassword` / `neo4j` | Auth |
| `GRAPH_IOC_STALE_DAYS` | `90` | Age threshold for “stale” isolated IOCs |
| `GRAPH_DELETE_STALE_ISOLATED_IOCS` | `0` | Set to `1` to `DETACH DELETE` stale isolated IOCs when not using `--dry-run` |

Other graph maintenance (export/pack/import) is documented in the root [README.md](../README.md) and [scrapers/README.md](../scrapers/README.md).
