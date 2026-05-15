# Threat Intel runtime (Docker Compose)

Veil runs as **three isolated layers** under [`deploy/`](../deploy/): **scrape** (fetch + ledger + NATS publish), **pipeline** (normalize → `ingest.>`), **graph** (Neo4j ingest + API). Default graph stack: **Neo4j** → **graph-bootstrap** → **HTTP API**. Full pipeline: [`./scripts/compose-up-full.sh`](../scripts/compose-up-full.sh) or merged compose files (see [deploy/README.md](../deploy/README.md)). Worker scaling: `PIPELINE_WORKER_SCALE`, `INGEST_WORKER_SCALE` in [deploy.md](deploy.md).

## Ports

| Service | Port | Notes |
|---------|------|--------|
| Neo4j Browser | `${NEO4J_HTTP_PORT:-7474}` (host) | Bolt `${NEO4J_BOLT_PORT:-7687}`; map with `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` if defaults are busy |
| HTTP API | 8090 | `API_PORT` to override published port |
| nginx LB | `${LB_HTTP_PORT:-8888}` | Optional; not in default compose — scale **`api`** via replicas manually if needed |
| Proxybroker | 8099 | Full stack only; `PROXYBROKER_PORT` |
| NATS client | `${NATS_CLIENT_PORT:-4222}` | Full stack (`compose-up-full`); maps container `4222` |
| NATS monitoring | `${NATS_MONITOR_PORT:-8222}` | HTTP on container `8222`; **`nats`** healthcheck uses **`http://127.0.0.1:8222/healthz`** |

## Build and environment checklist (repeatable builds)

Use this before a full scrape run or a reproducible graph pack.

1. **Full stack build** — `docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml build` (or `./scripts/compose-up-full.sh`). If builds fail with **EOF** or truncated module downloads, set **`GOPROXY`** for Docker builds or use a stable proxy.
2. **Bolt / Browser ports** — if `7687` or `7474` are taken, set **`NEO4J_BOLT_PORT`** / **`NEO4J_HTTP_PORT`** and stop duplicate Neo4j containers.
3. **Graph-only** — `docker compose up --build -d` then **`curl -sS http://localhost:${API_PORT:-8090}/health`**.
4. **Full pipeline** — `./scripts/compose-up-full.sh`; **`scrape_worker`** waits for **`nats` healthy**; **`ingest_worker`** waits for **Neo4j** and **NATS**. TLS/EOF during `go mod download` in Docker is usually network/proxy, not application code.

## Compose service reference

Layer compose files: [deploy/scrape/compose.yml](../deploy/scrape/compose.yml), [deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml), [deploy/graph/compose.yml](../deploy/graph/compose.yml). Root [docker-compose.yml](../docker-compose.yml) includes graph only. Optional local graph pack: [docker-compose.testpack.yml](../docker-compose.testpack.yml).

### neo4j

| | |
|--|--|
| **Profile** | *(none — starts with default `docker compose up`)* |
| **Image** | `neo4j:5` |
| **Purpose** | Graph database; APOC enabled for exports |
| **Volumes** | `neo4j_data`, `neo4j_logs`; host `./data/neo4j_user_export` → import path |
| **Health** | `cypher-shell` `RETURN 1` |
| **Env** | `NEO4J_AUTH`, `NEO4J_PLUGINS` (APOC), `NEO4J_apoc_export_file_enabled=true` for export |

### graph-bootstrap

| | |
|--|--|
| **Profile** | *(none)* |
| **Build** | [deploy/graph/docker/graph-bootstrap.Dockerfile](../deploy/graph/docker/graph-bootstrap.Dockerfile) |
| **Purpose** | One-shot import of a graph pack ZIP before API starts |
| **Depends on** | `neo4j` healthy |
| **Restart** | `no` |
| **Env** | `GRAPH_PACK_*`, `NEO4J_*` — see [Graph bootstrap (usage mode)](#graph-bootstrap-usage-mode) |

### api

| | |
|--|--|
| **Profile** | *(none)* |
| **Build** | [deploy/graph/docker/api.Dockerfile](../deploy/graph/docker/api.Dockerfile) |
| **Purpose** | Categorical REST API over Neo4j |
| **Ports** | `${API_PORT:-8090}:8090` |
| **Depends on** | `neo4j` healthy, `graph-bootstrap` completed |
| **Health** | `GET /health` |
| **Env** | `API_LISTEN_ADDR`, `NEO4J_*` |

### mcp

| | |
|--|--|
| **Profile** | `mcp` |
| **Build** | dev-only: build from [graph/mcp/](../graph/mcp/) (not in default `deploy/` compose) |
| **Purpose** | Stdio MCP tools (same queries as API) — [mcp/README.md](../mcp/README.md) |
| **Depends on** | `neo4j` healthy, `graph-bootstrap` completed |
| **Env** | `NEO4J_*` |

### nats (JetStream broker)

| | |
|--|--|
| **Compose** | [deploy/scrape/compose.yml](../deploy/scrape/compose.yml) (full stack) |
| **Image** | `nats:2.10-alpine` |
| **Command** | `-js -m 8222` (JetStream + monitoring) |
| **Purpose** | Two-stream bus: scrapers → **`scrape.>`** (`SCRAPE`); **pipeline_worker** → **`ingest.>`** (`INGEST`) |
| **Ports** | Client and monitoring (see [Ports](#ports)) |
| **Health** | `wget` on `http://127.0.0.1:8222/healthz` (JetStream monitoring) |
| **Streams** | **`SCRAPE`** (`scrape.>`), **`INGEST`** (`ingest.>`) — ensured by **scrapepub** / **pipeline_worker** / **ingest_worker** |

### pipeline_worker

| | |
|--|--|
| **Compose** | [deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml) |
| **Build** | [deploy/pipeline/docker/pipeline_worker.Dockerfile](../deploy/pipeline/docker/pipeline_worker.Dockerfile) |
| **Module** | [pipeline/pipeline_worker/](../pipeline/pipeline_worker/) |
| **Purpose** | Pull **`scrape.>`**, normalize/dedup **`scrapev1`** → **`ingestv1`**, publish **`ingest.>`** (per-domain subjects via `*_INGEST_SUBJECT`) |
| **Depends on** | **`nats` healthy** (when scrape layer is merged) |
| **Scale** | `PIPELINE_WORKER_SCALE` — shared durable `pipeline_worker` |
| **Env** | `NATS_URL`, `NATS_SCRAPE_STREAM`, `NATS_SCRAPE_DURABLE`, `NATS_SCRAPE_SUBSCRIBE_SUBJECT`, `PIPELINE_BATCH`, `PIPELINE_MAX_WAIT`, `DS_INGEST_SUBJECT`, `TI_INGEST_SUBJECT`, … |

### ingest_worker

| | |
|--|--|
| **Compose** | [deploy/graph/compose.yml](../deploy/graph/compose.yml) |
| **Build** | [deploy/graph/docker/ingest_worker.Dockerfile](../deploy/graph/docker/ingest_worker.Dockerfile) |
| **Module** | [graph/ingest_worker/README.md](../graph/ingest_worker/README.md) |
| **Purpose** | Long-running **JetStream pull consumer**: reads `ingestv1` from **`ingest.>`**, writes **Neo4j** (AppSec via `graph/storage/*`; ti/vuln/lola/ds via `graph/sources/*` + `graph/workeringest/*`) |
| **Depends on** | `neo4j` healthy, **`nats` healthy** (full stack) |
| **Restart** | `unless-stopped` |
| **Scale** | `INGEST_WORKER_SCALE` — shared durable `ingest_worker` |
| **Env** | `NEO4J_*`, `NATS_URL`, `NATS_INGEST_STREAM`, `NATS_DURABLE`, `NATS_SUBSCRIBE_SUBJECT`, `INGEST_BATCH`, `INGEST_MAX_WAIT` |

Use **`ingest_worker`** whenever scrape producers run; without it, messages stay in JetStream until drained.

### proxybroker

| | |
|--|--|
| **Compose** | [deploy/scrape/compose.yml](../deploy/scrape/compose.yml) |
| **Build** | [deploy/scrape/docker/proxybroker.Dockerfile](../deploy/scrape/docker/proxybroker.Dockerfile) |
| **Purpose** | HTTP proxy pool for scrapers (`*_PROXY_URLS`) |
| **Ports** | `${PROXYBROKER_PORT:-8099}:8099` |

### scrape_worker (factory orchestrator)

| | |
|--|--|
| **Compose** | [deploy/scrape/compose.yml](../deploy/scrape/compose.yml) |
| **Build** | [deploy/scrape/docker/scrape_worker.Dockerfile](../deploy/scrape/docker/scrape_worker.Dockerfile) |
| **Module** | [scrape/scrape_worker/](../scrape/scrape_worker/) |
| **Purpose** | Runs selected sources via [scrape/factory](../scrape/factory/); publishes **`scrapev1`** to **`scrape.>`** (batch job, exits 0) |
| **Depends on** | **`nats` healthy** |
| **Partition** | `SCRAPE_WORKER_PARTITION=1` → `scrape_worker_fast` + `scrape_worker_slow` ([deploy/compose.scale.yml](../deploy/compose.scale.yml)) |
| **Env** | **`SCRAPE_SOURCES`**, **`TI_FEEDS`**, **`TI_JSONL_FILE`**, **`SBOM_CVE_LIST_FILE`**, **`VITESS_DSN`**, per-source scrape subjects; optional `GITHUB_TOKEN` (rate limits only) |

Sources live under [scrape/sources/](../scrape/sources/). They publish **`scrapev1`** only (no Bolt). **`pipeline_worker`** → **`ingestv1`**; **`ingest_worker`** → Neo4j. See [scrape/README.md](../scrape/README.md).

### NATS subjects

| Variable | Default | Meaning |
|----------|---------|--------|
| `SCRAPE_SOURCES` | `ds,vuln,lola,ti,sbom,coderules,nuclei` | Comma-separated sources for **`scrape_worker`** |
| `TI_FEEDS` | `kev,urlhaus,threatfox,malwarebazaar,feodo` | Feed list when `ti` is enabled |
| `TI_JSONL_FILE` | `/app/example.jsonl` (Compose) | Optional JSONL path; empty to skip |
| `NATS_URL` | `nats://nats:4222` in Compose | NATS client URL |
| `VULN_SCRAPE_SUBJECT` | `scrape.vuln.events` | Scraper publish subject for **`vuln`** |
| `LOLA_SCRAPE_SUBJECT` | `scrape.lola.events` | Scraper publish for **`lola`** |
| `DS_SCRAPE_SUBJECT` | `scrape.ds.events` | Scraper publish for **`ds`** |
| `TI_SCRAPE_SUBJECT` | `scrape.ti.events` | Scraper publish for **`ti`** |
| `SBOM_SCRAPE_SUBJECT` | `scrape.appsec.sbom` | Scraper publish for `sbom` |
| `CODERULES_SCRAPE_SUBJECT` | `scrape.appsec.coderules` | Scraper publish for `coderules` |
| `NUCLEI_SCRAPE_SUBJECT` | `scrape.appsec.nuclei` | Scraper publish for `nuclei` |
| `DS_INGEST_SUBJECT` | `ingest.ds.events` | **pipeline_worker** publish for DS |
| `TI_INGEST_SUBJECT` | `ingest.ti.events` | pipeline publish for TI |
| `VULN_INGEST_SUBJECT` | `ingest.vuln.events` | pipeline publish for vuln |
| `LOLA_INGEST_SUBJECT` | `ingest.lola.events` | pipeline publish for lola |
| `SBOM_INGEST_SUBJECT` | `ingest.appsec.sbom` | pipeline publish for sbom |
| `CODERULES_INGEST_SUBJECT` | `ingest.appsec.coderules` | pipeline publish for coderules |
| `NUCLEI_INGEST_SUBJECT` | `ingest.appsec.nuclei` | pipeline publish for nuclei |
| `SBOM_CVE_LIST_FILE` | `/fixtures/cve_list_seed.txt` in Compose | CVE list for OSV (one `CVE-…` per line; `#` comments allowed) |
| `SBOM_CVE_LIST_URL` | empty | Alternative CVE list URL if file unset |
| `VITESS_DSN` / `MYSQL_DSN` | `veil:veilpass@tcp(crawl-db:3306)/veil_ledger` in scrape compose | Crawl ledger ([scrape/ledger](../scrape/ledger/)); records URL + `content_sha256` |
| `SCRAPE_MIN_REFETCH_AFTER` | `24h` | Min refetch interval (`periodic` policy) |
| `SCRAPE_FORCE_REFETCH` | `0` | `1` = ignore ledger (full refetch) |
| `LOFTS_SKIP_ON_ERROR` | unset | `true` = LOFTS fetch errors are warnings (do not fail `lola`) |
| `SCRAPE_FAIL_FAST` | unset | `1` = stop all sources on first source error |
| `NATS_INGEST_STREAM` | `INGEST` | Stream name (worker) |
| `NATS_DURABLE` | `ingest_worker` | Durable consumer name |
| `NATS_SUBSCRIBE_SUBJECT` | `ingest.>` | Worker pull filter (AppSec, TI, vuln, lola, ds, …) |
| `INGEST_BATCH` | `10` | Max messages per fetch |
| `INGEST_MAX_WAIT` | `5s` | Fetch wait |

JetStream dedup: **`Nats-Msg-Id`** from envelope **`idempotency_key`** ([pipeline/pub](../pipeline/pub), [pipeline/contract/ingestv1](../pipeline/contract/ingestv1)).

Contract details: [docs/ingest-contract.md](ingest-contract.md).

## Graph bootstrap (usage mode)

Init container `graph-bootstrap` runs once after Neo4j is healthy. Pack resolution order:

1. Bind mount **`/pack/host.zip`** inside the container (optional; add under `graph-bootstrap.volumes` in a local override file).
2. **`GRAPH_PACK_FILE`** — path *inside the container* to a `.zip` (if you extend the service with a volume).
3. **`GRAPH_PACK_URL`** — HTTP(S) URL to a pack ZIP.
4. If **`GRAPH_PACK_DEFAULT=1`** (default): download **`GRAPH_PACK_DEFAULT_URL`**, or the built-in default release asset if unset.

Skip import entirely: **`GRAPH_PACK_SKIP=1`**.

| Variable | Default | Meaning |
|----------|---------|--------|
| `GRAPH_PACK_SKIP` | `0` | `1` = exit 0 without importing |
| `GRAPH_PACK_DEFAULT` | `1` | `0` = do not download the default release ZIP when no file/URL |
| `GRAPH_PACK_URL` | empty | HTTP(S) URL of the pack ZIP |
| `GRAPH_PACK_DEFAULT_URL` | built-in GitHub asset | Overrides the default download URL when `GRAPH_PACK_DEFAULT=1` |
| `GRAPH_PACK_FILE` | empty | Path **inside** the bootstrap container (mount a volume if needed) |

Compose passes these from the host for `graph-bootstrap` (see [docker-compose.yml](../docker-compose.yml) `environment`).

Checksum: `manifest.json` `sha256` must match `graph.cypher` (same rules as [scripts/import-graph-pack.sh](../scripts/import-graph-pack.sh)).

## Full scrape stack

```bash
./scripts/compose-up-full.sh
# or: docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml up --build
```

Starts **`proxybroker`**, **`crawl-db`**, **`nats`**, **`pipeline_worker`**, **`ingest_worker`**, **`scrape_worker`** (default `SCRAPE_SOURCES=ds,vuln,lola,ti,sbom,coderules,nuclei`). A **second** scrape run should log skipped/unchanged feeds. TI normalization runs in **`pipeline_worker`**. Scale consumers: `PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/compose-up-full.sh`.

Verify ledger: `docker compose exec crawl-db mysql -uveil -pveilpass veil_ledger -e 'SELECT resource_key, fetch_policy, last_fetched_at FROM crawl_resource LIMIT 20;'`

After data is in Neo4j, export a pack from the host:

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.2 ./scripts/build-graph-pack.sh
```

### Fast-rich graph pack profile (~25 min)

For a **richer** pack than smoke/seed without a full NVD crawl, use [scripts/graph-pack-run-v032.sh](../scripts/graph-pack-run-v032.sh): all seven `SCRAPE_SOURCES`, `NVD_MAX_PAGES=1` (~2000 CVE), no Atomic/Metasploit bulk, boosted Sigma/YARA/OSV/GHSA/coderules/nuclei limits. NVD **CWE/CPE enrichment** flows through shared [pkg/nvdparse](../pkg/nvdparse) (pipeline `vulnFromNVDPage` → graph `HAS_CWE` / `AFFECTS`).

Resilience env (set in the script and [deploy/scrape/compose.yml](../deploy/scrape/compose.yml)):

| Variable | Purpose |
|----------|---------|
| `LOFTS_SKIP_ON_ERROR=true` | LOFTS DNS failures do not abort `lola` |
| `SCRAPE_FAIL_FAST=1` | Stop on first source error (default: continue other sources, exit 1 if any failed) |

```bash
./scripts/graph-pack-run-v032.sh
# after scrape_worker finishes and ingest drains:
./scripts/graph-dedup-cleanup.sh --dry-run && ./scripts/graph-dedup-cleanup.sh
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.2 ./scripts/build-graph-pack.sh
```

Verify NVD enrichment in Neo4j:

```bash
./scripts/verify-nvd-enrichment.sh
```

Neo4j must have `NEO4J_apoc_export_file_enabled=true` (set in [deploy/graph/compose.yml](../deploy/graph/compose.yml)).

### Smoke checklist

1. **Default stack (no scrape):** `docker compose up --build -d` → wait for **`api` healthy** → `curl` **`/health`** and a few **`/v1/...`** calls (see [curl examples](#curl-examples)).
2. **Graph pack without GitHub download:** place **`threat-intel-graph-v0.3.2.zip`** under `data/neo4j_user_export/releases/` and run `docker compose -f docker-compose.yml -f docker-compose.testpack.yml up --build -d` (see [docker-compose.testpack.yml](../docker-compose.testpack.yml)).
3. **Scrape + NATS:** `./scripts/smoke_scrape_e2e.sh --up`; confirm JetStream drains and Neo4j gains nodes (see [scrape/README.md](../scrape/README.md), [graph/README.md](../graph/README.md)).
4. **Release asset:** the default URL in [deploy/graph/docker/graph-bootstrap.sh](../deploy/graph/docker/graph-bootstrap.sh) must point at a ZIP that contains **`manifest.json`** + **`graph.cypher`** with matching **`sha256`**. Bump version and URLs if the dump changes.

### E2E scrape smoke (slice 8 v2)

Automated checks for the full scrape → pipeline → ingest path:

```bash
./scripts/smoke_scrape_e2e.sh --up          # compose up scrape stack, wait for scrape_worker exit
./scripts/smoke_scrape_e2e.sh               # NATS health, crawl_resource rows, Neo4j counts, API /health
./scripts/smoke_scrape_e2e.sh --restart-scrape   # pass 2: ledger unchanged / skip publish
```

Env overrides: `SCRAPE_SVC`, `PIPELINE_SVC`, `INGEST_SVC`, `PIPELINE_WORKER_SCALE`, `INGEST_WORKER_SCALE`, `SCRAPE_WORKER_PARTITION`, `NATS_MON`, `API_URL`, `CRAWL_MYSQL`, `SMOKE_WAIT_SEC`.

## HTTP API (categorical)

Base URL: `http://localhost:8090` (with default compose).

- `GET /health`
- `GET /v1/categories` — product categories (`vuln`, `ti`, `detection`, `lola`, `mitre`, `sbom`, `code_rules`, `dast`) and Neo4j label sets.
- `GET /v1/categories/{category}/kinds` — labels present in the graph with counts.
- `GET /v1/categories/{category}/nodes?kind=Vulnerability&limit=50&offset=0`
- `GET /v1/categories/{category}/search?q=cve&kind=&limit=50`
- `GET /v1/nodes/{id}` — elementId or `id` / `cve` / `uri` / `link`
- `GET /v1/nodes/{id}/neighbors?depth=1&limit=500`
- `GET /v1/kinds` — all distinct labels (legacy discovery)

OpenAPI sketch: [openapi.yaml](openapi.yaml).

### curl examples

Assuming default `API_PORT=8090` after `docker compose up`:

```bash
curl -sS http://localhost:8090/health
curl -sS http://localhost:8090/v1/categories | jq .
curl -sS 'http://localhost:8090/v1/categories/vuln/kinds' | jq .
curl -sS 'http://localhost:8090/v1/categories/vuln/nodes?kind=Vulnerability&limit=5' | jq .
curl -sS 'http://localhost:8090/v1/categories/vuln/search?q=cve&limit=5' | jq .
curl -sS 'http://localhost:8090/v1/kinds' | jq .
```

Replace `vuln` / `Vulnerability` with other [categories](../graph/neo4jclient/query/categories.go) and labels as needed.

## MCP (stdio)

Same categorical logic as the API. Module: [graph/mcp/](../graph/mcp/). Run against a running Neo4j (after bootstrap or scrape).

From source:

```bash
cd graph/mcp && go run ./cmd
```

Category-first tools: `ti_list_categories`, `ti_list_kinds_in_category`, `ti_nodes_by_category`, `ti_search_in_category`; legacy tools remain for raw label access.

## Optional: local pack file

**Quick smoke test with the repo’s sample pack** (no download):

```bash
docker compose -f docker-compose.yml -f docker-compose.testpack.yml up --build -d
```

See [docker-compose.testpack.yml](../docker-compose.testpack.yml) (bind-mounts `data/neo4j_user_export/releases/threat-intel-graph-v0.3.2.zip` as `/pack/host.zip` and sets `GRAPH_PACK_DEFAULT=0`).

Re-importing the same pack into **non-empty** Neo4j (existing constraints) will fail. For a clean ZIP import use `docker compose … down -v` or import with **`ingest_worker` stopped** (it may create constraints before bootstrap). Testpack flow: start **`neo4j`**, run **`graph-bootstrap`** once (`docker compose run --rm graph-bootstrap`), then **`GRAPH_PACK_SKIP=1 docker compose up api -d`**.

Create `docker-compose.override.yml` (gitignored by convention or not committed) with:

```yaml
services:
  graph-bootstrap:
    volumes:
      - ./my-pack.zip:/pack/host.zip:ro
```

The bootstrap script copies `/pack/host.zip` when it exists and is non-empty.
