# Veil deploy (per layer)

**Compose stack presets** (named overlay chains + profiles): [stacks/](stacks/README.md).

| Layer | Compose | Services |
|-------|---------|----------|
| Scrape | [scrape/compose.yml](scrape/compose.yml) | `crawl-db`, `nats`, `scrape_worker`, `proxybroker` |
| Pipeline | [pipeline/compose.yml](pipeline/compose.yml) | `pipeline_worker` |
| Graph | [graph/compose.yml](graph/compose.yml) | `neo4j`, `graph-bootstrap`, `ingest_worker`, `api` |
| Engage | [engage/compose.yml](engage/compose.yml) | `engage-api`, `engage-mcp`, `engage-worker`, `engage-runner` (profile `runner`; opt-in offensive tools) |

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

## Hybrid deploy (P5 — Terraform + Ansible + Helm)

Production model: **TF** provisions cloud → **Ansible** data plane (Compose) → **Helm** control plane (K8s). See [docs/deploy-platform-hybrid.md](../docs/deploy-platform-hybrid.md) and [.cursor/plans/veil_deploy_platform_p5_hybrid.plan.md](../.cursor/plans/veil_deploy_platform_p5_hybrid.plan.md).

| Path | Role |
|------|------|
| [terraform/](terraform/README.md) | Infra + compose env generation |
| [ansible/](ansible/README.md) | VM configure, stateful stack, scrape cron |
| [helm/veil/](helm/veil/README.md) | api, engage-api, workers HPA, scrape CronJob |

## Terraform (local IaC)

Declarative Compose env and optional managed `up`/`down`: [terraform/README.md](terraform/README.md).

```bash
cd deploy/terraform/environments/local && terraform init && terraform apply
```

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

**Development (default)** — Neo4j and API ports on the host:

| Service | Host port |
|---------|-----------|
| Neo4j HTTP | `7474` |
| Neo4j Bolt | `7687` |
| API | `8090` |
| MCP HTTP (`--profile mcp`) | `8091` |

```bash
docker compose up --build -d
# or: docker compose -f deploy/graph/compose.yml up --build -d
```

Root [docker-compose.yml](../docker-compose.yml) includes only the graph layer.

**Graph read smoke (no scrape/NATS):**

```bash
make test-graph-read-smoke
# or: ./scripts/test/smoke-graph-read.sh --up
```

**Production secure overlay** — only nginx `443` on the host; see [docs/deploy-secure.md](../docs/deploy-secure.md):

```bash
docker compose -f deploy/graph/compose.yml -f deploy/graph/compose.secure.yml \
  --profile mcp --env-file deploy/profiles/secure-graph.env up -d --build
```

## Graph pack releases

Naming: ZIP **`veil-graph-vX.Y.Z.zip`**, GitHub tag **`veil-graph-vX.Y.Z`**. See [docs/graph-pack.md](../docs/graph-pack.md).

| Release | Notes |
|---------|--------|
| [veil-graph-v0.4.5](https://github.com/butbeautifulv/veil/releases/tag/veil-graph-v0.4.5) | Target format on `main` (publish when built) |
| [v0.3.2-graph-pack](https://github.com/butbeautifulv/veil/releases/tag/v0.3.2-graph-pack) | Legacy `threat-intel-graph-v0.3.2.zip` (redirects) |

Build (incremental crawl state in `var/veil/`):

```bash
./scripts/graph-pack/profile-incremental-pack.sh   # or profile-fast-rich.sh / --full
./scripts/housekeeping/graph-dedup-cleanup.sh
./scripts/graph-pack/export-cypher.sh
GRAPH_PACK_VERSION=v0.4.5 ./scripts/graph-pack/build.sh
GRAPH_PACK_VERSION=v0.4.5 ./scripts/release/publish-graph-pack.sh --skip-build
```

Script index: [scripts/README.md](../scripts/README.md).

## Engage layer (opt-in)

Offensive tooling is **not** part of the default full stack. Use a separate compose file:

```bash
docker compose -f deploy/engage/compose.yml up -d --build engage-api engage-mcp
```

| Service | Host port (dev) | Role |
|---------|-----------------|------|
| engage-api | 8890 | REST + workflows + async jobs |
| engage-mcp | 8892 | Streamable HTTP MCP (`ENGAGE_MCP_HTTP_ENABLED=1`) |
| engage-worker | — | Background job processor |
| engage-runner | none | Toolbox image; `--profile runner` |

Overlays: [compose.runner.yml](engage/compose.runner.yml) (docker exec), [compose.queue.yml](engage/compose.queue.yml) (Redis jobs), [compose.nats.yml](engage/compose.nats.yml) (NATS jobs), [compose.events.yml](engage/compose.events.yml) (standalone NATS + `engage-events-worker`; profile **`graph-ingest`** adds Neo4j + `ingest_worker`), [compose.veil-stack.yml](engage/compose.veil-stack.yml) (engage on **shared** Veil NATS/Neo4j — use with `./scripts/ops/compose-up-veil-engage.sh`, not together with `compose.events.yml`).

Secure overlay: [engage/compose.secure.yml](engage/compose.secure.yml) + [profiles/secure-engage.env](profiles/secure-engage.env).

Runtime: [docs/engage-runtime.md](../docs/engage-runtime.md). Catalog: [docs/engage-tools.md](../docs/engage-tools.md).
