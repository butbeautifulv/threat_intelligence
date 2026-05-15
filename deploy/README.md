# Veil deploy (per layer)

| Layer | Compose | Services |
|-------|---------|----------|
| Scrape | [scrape/compose.yml](scrape/compose.yml) | `crawl-db`, `nats`, `scrape_worker`, `proxybroker` |
| Pipeline | [pipeline/compose.yml](pipeline/compose.yml) | `pipeline_worker` |
| Graph | [graph/compose.yml](graph/compose.yml) | `neo4j`, `graph-bootstrap`, `ingest_worker`, `api` |

## Full stack

```bash
./scripts/compose-up-full.sh
```

Equivalent:

```bash
docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml up --build
```

Docker build context is the repository root; each Dockerfile copies only its layer tree (`scrape/`, `pipeline/`, or `graph/`).

## Worker scaling (parallel NATS consumers)

`pipeline_worker` and `ingest_worker` use JetStream pull consumers with a **shared durable name**. Multiple replicas compete for messages (safe scale-out).

| Variable | Default | Meaning |
|----------|---------|--------|
| `PIPELINE_WORKER_SCALE` | `1` | Replicas of `pipeline_worker` |
| `INGEST_WORKER_SCALE` | `1` | Replicas of `ingest_worker` |
| `SCRAPE_WORKER_PARTITION` | `0` | `1` = two scrape jobs (`scrape_worker_fast` + `scrape_worker_slow`) instead of one `scrape_worker` |

Examples:

```bash
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/compose-up-full.sh

SCRAPE_WORKER_PARTITION=1 ./scripts/compose-up-full.sh
```

Manual scale (without the script):

```bash
docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml \
  up -d --scale pipeline_worker=3 --scale ingest_worker=3
```

**Do not** scale `scrape_worker` with the same `SCRAPE_SOURCES` on multiple replicas (duplicate crawl). Use `SCRAPE_WORKER_PARTITION=1` or split `SCRAPE_SOURCES` per container.

Partition overlay: [compose.scale.yml](compose.scale.yml) (`scrape_worker_fast`: ti,sbom,coderules,nuclei; `scrape_worker_slow`: ds,vuln,lola).

## E2E smoke

```bash
./scripts/smoke_scrape_e2e.sh --up
./scripts/smoke_scrape_e2e.sh --restart-scrape   # ledger pass 2
```

Smoke defaults: `SCRAPE_SOURCES=ti,sbom` (minimal), `NVD_MAX_PAGES=1`, `GRAPH_PACK_SKIP=1`, `SMOKE_CLEAN_VOLUMES=1`. Full crawl: `SCRAPE_SOURCES=ds,vuln,lola,ti,sbom,coderules,nuclei`. Optional `GITHUB_TOKEN` raises GitHub API rate limits only.

```bash
GRAPH_PACK_SKIP=0 ./scripts/smoke_scrape_e2e.sh --up
```

With scaled workers:

```bash
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/smoke_scrape_e2e.sh --up
```

## Docker builds

Images use layer `go.work` (`GOWORK=/build/<layer>/go.work`), not root `go.work`. Build context is the **repository root**.

## Graph-only (API + Neo4j)

```bash
docker compose up --build -d
# or: docker compose -f deploy/graph/compose.yml up --build -d
```

Root [docker-compose.yml](../docker-compose.yml) includes only the graph layer.

## Graph pack releases

| Release | Notes |
|---------|--------|
| [v0.3.2-graph-pack](https://github.com/butbeautifulv/threat_intelligence/releases/tag/v0.3.2-graph-pack) | Fast-rich partial crawl; **without** pipeline CWE/CPE fix |
| Current `main` | Use [pipeline/pkg/nvd/parse](../pipeline/pkg/nvd/parse/) + rebuild for full `HAS_CWE` / `AFFECTS` |

Build:

```bash
./scripts/graph-pack-run-v032.sh   # or compose-up-full with env overrides
./scripts/graph-dedup-cleanup.sh
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.2 ./scripts/build-graph-pack.sh
```

Runtime details: [docs/threatintel-runtime.md](../docs/threatintel-runtime.md). Script reference: [scripts/README.md](../scripts/README.md).
