# Veil deploy (per layer)

| Layer | Compose | Services |
|-------|---------|----------|
| Scrape | [scrape/compose.yml](scrape/compose.yml) | `crawl-db`, `nats`, `scrape_worker`, `proxybroker` |
| Pipeline | [pipeline/compose.yml](pipeline/compose.yml) | `pipeline_worker` |
| Graph | [graph/compose.yml](graph/compose.yml) | `neo4j`, `graph-bootstrap`, `ingest_worker`, `api` |

## Full stack

```bash
./scripts/ops/compose-up-full.sh
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
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/ops/compose-up-full.sh

SCRAPE_WORKER_PARTITION=1 ./scripts/ops/compose-up-full.sh
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
./scripts/test/smoke-scrape-e2e.sh --up
./scripts/test/smoke-scrape-e2e.sh --restart-scrape   # ledger pass 2
```

Smoke defaults: `SCRAPE_SOURCES=ti,sbom` (minimal), `NVD_MAX_PAGES=1`, `GRAPH_PACK_SKIP=1`, `SMOKE_CLEAN_VOLUMES=1`. Full crawl: `SCRAPE_SOURCES=ds,vuln,lola,ti,sbom,coderules,nuclei`. Optional `GITHUB_TOKEN` raises GitHub API rate limits only.

```bash
GRAPH_PACK_SKIP=0 ./scripts/test/smoke-scrape-e2e.sh --up
```

With scaled workers:

```bash
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/test/smoke-scrape-e2e.sh --up
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

Naming: ZIP **`veil-graph-vX.Y.Z.zip`**, GitHub tag **`veil-graph-vX.Y.Z`**. See [docs/graph-pack.md](../docs/graph-pack.md).

| Release | Notes |
|---------|--------|
| [veil-graph-v0.4.2](https://github.com/butbeautifulv/veil/releases/tag/veil-graph-v0.4.2) | Target format on `main` (publish when built) |
| [v0.3.2-graph-pack](https://github.com/butbeautifulv/veil/releases/tag/v0.3.2-graph-pack) | Legacy `threat-intel-graph-v0.3.2.zip` (redirects) |

Build (incremental crawl state in `var/veil/`):

```bash
./scripts/graph-pack/profile-incremental-pack.sh   # or profile-fast-rich.sh / --full
./scripts/housekeeping/graph-dedup-cleanup.sh
./scripts/graph-pack/export-cypher.sh
GRAPH_PACK_VERSION=v0.4.2 ./scripts/graph-pack/build.sh
GRAPH_PACK_VERSION=v0.4.2 ./scripts/release/publish-graph-pack.sh --skip-build
```

Script index: [scripts/README.md](../scripts/README.md).
