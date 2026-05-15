# Deploy and scale

Per-layer compose: [deploy/README.md](../deploy/README.md). Runtime details: [threatintel-runtime.md](threatintel-runtime.md).

## Full stack

```bash
./scripts/compose-up-full.sh
```

Equivalent:

```bash
docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml up --build
```

Docker **build context** is the repository root; Dockerfiles copy `pkg/`, `scrape/`, `pipeline/`, or `graph/` as needed.

## Scale pipeline and ingest workers

JetStream pull consumers share a **durable name**; multiple replicas compete for messages (safe scale-out).

```bash
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/compose-up-full.sh
```

Do **not** scale `scrape_worker` with the same `SCRAPE_SOURCES` on multiple replicas. Use `SCRAPE_WORKER_PARTITION=1` ([deploy/compose.scale.yml](../deploy/compose.scale.yml)) or split sources per container.

## Graph-only (API + Neo4j)

```bash
docker compose up --build -d
# or: docker compose -f deploy/graph/compose.yml up --build -d
```

Root [docker-compose.yml](../docker-compose.yml) includes only the graph layer.

## E2E smoke

```bash
./scripts/smoke_scrape_e2e.sh --up
./scripts/smoke_scrape_e2e.sh --restart-scrape
```

Minimal scrape (default): `SCRAPE_SOURCES=ti,sbom`. Full crawl: `SCRAPE_SOURCES=ds,vuln,lola,ti,sbom,coderules,nuclei`.

## Graph pack releases

| Release | Notes |
|---------|--------|
| [v0.3.2-graph-pack](https://github.com/butbeautifulv/threat_intelligence/releases/tag/v0.3.2-graph-pack) | Fast-rich partial crawl; **without** pipeline CWE/CPE fix |
| Current `main` | Use [pkg/nvdparse](../pkg/nvdparse/) + rebuild for full `HAS_CWE` / `AFFECTS` |

Build:

```bash
./scripts/graph-pack-run-v032.sh   # or compose-up-full with env overrides
./scripts/graph-dedup-cleanup.sh
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.2 ./scripts/build-graph-pack.sh
```

See [threatintel-runtime.md](threatintel-runtime.md) and [scripts/build-graph-pack.sh](../scripts/build-graph-pack.sh).
