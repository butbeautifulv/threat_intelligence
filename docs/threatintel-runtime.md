# Threat Intel runtime (Docker Compose)

Default stack: **Neo4j** → **graph-bootstrap** (import graph pack) → **HTTP API**. Live scraping, **NATS**, **ingest-worker**, and **proxybroker** are opt-in via the **`scrape`** profile. **MCP** uses **`profiles: ["mcp"]`**. Optional **nginx** load balancer in front of the API: profile **`deploy`** in [docker-compose.deploy.yml](../docker-compose.deploy.yml) — see [docs/deploy.md](deploy.md).

## Ports

| Service | Port | Notes |
|---------|------|--------|
| Neo4j Browser | `${NEO4J_HTTP_PORT:-7474}` (host) | Bolt `${NEO4J_BOLT_PORT:-7687}`; map with `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` if defaults are busy |
| HTTP API | 8090 | `API_PORT` to override published port |
| nginx LB | `${LB_HTTP_PORT:-8888}` | Profile **`deploy`** + [docker-compose.deploy.yml](../docker-compose.deploy.yml); HTTP to **`api`** replicas (see [docs/deploy.md](deploy.md)) |
| Proxybroker | 8099 | Only with `--profile scrape`; `PROXYBROKER_PORT` |
| NATS client | `${NATS_CLIENT_PORT:-4222}` | Only with `--profile scrape`; maps container `4222` |
| NATS monitoring | `${NATS_MONITOR_PORT:-8222}` | HTTP on container `8222`; **`nats`** healthcheck uses **`http://127.0.0.1:8222/healthz`** |

## Build and environment checklist (repeatable builds)

Use this before relying on a full **`--profile scrape`** run or a reproducible graph pack.

1. **`docker compose --profile scrape build`** — may run many `go mod download` steps inside Docker. If builds fail with **EOF** or truncated module downloads, set **`GOPROXY`** (and `GOSUMDB` if required) for Docker builds or use a stable corporate proxy; an unstable VPN often shows up here first.
2. **Bolt / Browser ports** — if `7687` or `7474` are taken on the host, set **`NEO4J_BOLT_PORT`** / **`NEO4J_HTTP_PORT`** (see [Ports](#ports)) and stop duplicate Neo4j containers.
3. **After `docker compose up --build`** — `cypher-shell` against the published Bolt URL; **`curl -sS http://localhost:${API_PORT:-8090}/health`** (see [HTTP API](#http-api-categorical)).
4. **Scrape profile** — once **`docker compose --profile scrape build`** succeeds, `docker compose --profile scrape up -d` is the baseline smoke; scrape services wait for **Neo4j healthy** and **NATS healthy** (`healthz` on the monitoring port). **`go mod download`** failures with **TLS handshake timeout** or **EOF** inside Docker are network/proxy problems (try **`GOPROXY`**, stable VPN, or retry), not necessarily code defects.

## Compose service reference

All definitions live in [docker-compose.yml](../docker-compose.yml). Optional NATS-only producer env: [docker-compose.scrape-nats.yml](../docker-compose.scrape-nats.yml). Optional API load balancer: [docker-compose.deploy.yml](../docker-compose.deploy.yml). **Binary / module docs:** see links in each subsection.

### neo4j

| | |
|--|--|
| **Profile** | *(none — starts with default `docker compose up`)* |
| **Image** | `neo4j:5` |
| **Purpose** | Graph database; APOC enabled for exports |
| **Volumes** | `neo4j_data`, `neo4j_logs`; host `./data/neo4j_user_export` → import path |
| **Health** | `cypher-shell` `RETURN 1` |
| **Env** | `NEO4J_AUTH`, memory, plugins (see compose file) |

### graph-bootstrap

| | |
|--|--|
| **Profile** | *(none)* |
| **Build** | [docker/graph-bootstrap.Dockerfile](../docker/graph-bootstrap.Dockerfile) |
| **Purpose** | One-shot import of a graph pack ZIP before API starts |
| **Depends on** | `neo4j` healthy |
| **Restart** | `no` |
| **Env** | `GRAPH_PACK_*`, `NEO4J_*` — see [Graph bootstrap (usage mode)](#graph-bootstrap-usage-mode) |

### api

| | |
|--|--|
| **Profile** | *(none)* |
| **Build** | [docker/api.Dockerfile](../docker/api.Dockerfile) |
| **Purpose** | Categorical REST API over Neo4j |
| **Ports** | `${API_PORT:-8090}:8090` |
| **Depends on** | `neo4j` healthy, `graph-bootstrap` completed |
| **Health** | `GET /health` |
| **Env** | `API_LISTEN_ADDR`, `NEO4J_*` |

### mcp

| | |
|--|--|
| **Profile** | `mcp` |
| **Build** | [docker/mcp.Dockerfile](../docker/mcp.Dockerfile) |
| **Purpose** | Stdio MCP tools (same queries as API) — [mcp/README.md](../mcp/README.md) |
| **Depends on** | `neo4j` healthy, `graph-bootstrap` completed |
| **Env** | `NEO4J_*` |

### nats (JetStream broker)

| | |
|--|--|
| **Profile** | `scrape` |
| **Image** | `nats:2.10-alpine` |
| **Command** | `-js -m 8222` (JetStream + monitoring) |
| **Purpose** | Message bus for optional **`INGEST_MODE=nats`** on `sbom`, `coderules`, `nuclei`, **`ti`**, **`vuln`**, **`lola`**, **`ds`** |
| **Ports** | Client and monitoring (see [Ports](#ports)) |
| **Health** | `wget` on `http://127.0.0.1:8222/healthz` (JetStream monitoring) |
| **Stream** | Created by publishers or **`ingest-worker`**: name **`INGEST`**, subjects **`ingest.>`** |

### ingest-worker

| | |
|--|--|
| **Profile** | `scrape` |
| **Build** | [docker/ingest-worker.Dockerfile](../docker/ingest-worker.Dockerfile) |
| **Module** | [scrapers/ingest-worker/README.md](../scrapers/ingest-worker/README.md) |
| **Purpose** | Long-running **JetStream pull consumer**: reads `ingestv1` envelopes from **`ingest.>`**, writes **Neo4j** using the same `MERGE` paths as **`direct`** scrapers (AppSec, **`ti`**, **`vuln`**, **`lola`**, **`ds`**) |
| **Depends on** | `neo4j` healthy, **`nats` healthy** |
| **Restart** | `on-failure` |
| **Env** | `NEO4J_*`, `NATS_URL` (Compose: `nats://nats:4222`), `NATS_INGEST_STREAM`, `NATS_DURABLE`, `NATS_SUBSCRIBE_SUBJECT`, `INGEST_BATCH`, `INGEST_MAX_WAIT` — full table in [scrapers/ingest-worker/README.md](../scrapers/ingest-worker/README.md) |

Use **`ingest-worker`** whenever any scraper publishes with **`INGEST_MODE=nats`**; otherwise messages stay in JetStream until drained.

### proxybroker

| | |
|--|--|
| **Profile** | `scrape` |
| **Build** | [docker/proxybroker.Dockerfile](../docker/proxybroker.Dockerfile) |
| **Purpose** | HTTP proxy pool for scrapers (`*_PROXY_URLS`) |
| **Ports** | `${PROXYBROKER_PORT:-8099}:8099` |

### Scrape ingest services (Neo4j writers or NATS publishers)

| Compose service | Dockerfile | Notes |
|-----------------|------------|--------|
| `vuln` | [docker/vuln.Dockerfile](../docker/vuln.Dockerfile) | NVD, Metasploit, Exploit-DB, optional Vulners; **`INGEST_MODE`**, **`VULN_NATS_SUBJECT`**; volume `data/cache`; depends on Neo4j + **`nats` healthy** |
| `lola` | [docker/lola.Dockerfile](../docker/lola.Dockerfile) | LOLBAS, GTFOBins, LOFTS, MITRE STIX; **`INGEST_MODE`**, **`LOLA_NATS_SUBJECT`**; depends on Neo4j + **`nats` healthy** |
| `ds` | [docker/ds.Dockerfile](../docker/ds.Dockerfile) | Sigma, YARA, Atomic, Caldera; **`INGEST_MODE`**, **`DS_NATS_SUBJECT`**; depends on Neo4j + **`nats` healthy** |
| `ti` | [docker/ti.Dockerfile](../docker/ti.Dockerfile) | KEV, URLhaus, ThreatFox, …; **`INGEST_MODE`**, **`TI_NATS_SUBJECT`**; depends on Neo4j + **`nats` healthy** |
| `sbom` | [docker/sbom.Dockerfile](../docker/sbom.Dockerfile) | OSV + GHSA; **`INGEST_MODE`**, **`SBOM_NATS_SUBJECT`**; depends on Neo4j + **`nats` healthy** |
| `coderules` | [docker/coderules.Dockerfile](../docker/coderules.Dockerfile) | CWE, Semgrep, CodeQL; **`INGEST_MODE`**, **`CODERULES_NATS_SUBJECT`**; depends on Neo4j + **`nats` healthy** |
| `nuclei` | [docker/nuclei.Dockerfile](../docker/nuclei.Dockerfile) | Nuclei templates; **`INGEST_MODE`**, **`NUCLEI_NATS_SUBJECT`**; depends on Neo4j + **`nats` healthy** |

All scrape ingest rows above use **`NEO4J_URI=neo4j://neo4j:7687`** (except **`coderules` / `nuclei` / `sbom` / `ti` / `vuln` / `lola` / `ds`** do not open Neo4j when **`INGEST_MODE=nats`**; **`sbom`** uses **`SBOM_CVE_LIST_FILE`** or **`SBOM_CVE_LIST_URL`** for OSV CVE ids in `nats` mode). See [scrapers/README.md](../scrapers/README.md) for per-feed env vars.

### NATS publish and consume (`INGEST_MODE`)

| Variable | Default | Meaning |
|----------|---------|--------|
| `INGEST_MODE` | `direct` | `direct` = scraper writes Neo4j; `nats` = publish envelopes (opt-in per service in Compose) |
| `NATS_URL` | `nats://nats:4222` in Compose | NATS client URL for publishers and **ingest-worker** |
| `SBOM_NATS_SUBJECT` | `ingest.appsec.sbom` | Publish subject for `sbom` |
| `SBOM_CVE_LIST_FILE` | empty (Compose scrape sets image default) | CVE list for OSV when **`INGEST_MODE=nats`** |
| `SBOM_CVE_LIST_URL` | empty | Alternative CVE list URL if file unset |
| `CODERULES_NATS_SUBJECT` | `ingest.appsec.coderules` | Publish subject for `coderules` |
| `NUCLEI_NATS_SUBJECT` | `ingest.appsec.nuclei` | Publish subject for `nuclei` |
| `TI_NATS_SUBJECT` | `ingest.ti.events` | Publish subject for **`ti`** when **`INGEST_MODE=nats`** |
| `VULN_NATS_SUBJECT` | `ingest.vuln.events` | Publish subject for **`vuln`** when **`INGEST_MODE=nats`** |
| `LOLA_NATS_SUBJECT` | `ingest.lola.events` | Publish subject for **`lola`** when **`INGEST_MODE=nats`** |
| `DS_NATS_SUBJECT` | `ingest.ds.events` | Publish subject for **`ds`** when **`INGEST_MODE=nats`** |
| `NATS_INGEST_STREAM` | `INGEST` | Stream name (worker) |
| `NATS_DURABLE` | `ingest-worker` | Durable consumer name |
| `NATS_SUBSCRIBE_SUBJECT` | `ingest.>` | Worker pull filter (AppSec, TI, vuln, lola, ds, …) |
| `INGEST_BATCH` | `10` | Max messages per fetch |
| `INGEST_MAX_WAIT` | `5s` | Fetch wait |

JetStream dedup: **`Nats-Msg-Id`** from envelope **`idempotency_key`** ([scrapers/ingestpub](../scrapers/ingestpub), [pkg/ingestv1](../pkg/ingestv1)).

### NATS-only producers (optional override)

When **`INGEST_MODE=nats`**, scrapers do not need Bolt credentials in the container. Use the extra file [docker-compose.scrape-nats.yml](../docker-compose.scrape-nats.yml) so **`NEO4J_*`** are unset (`null`) on **`vuln`**, **`sbom`**, **`lola`**, **`ds`**, **`ti`**, **`coderules`**, **`nuclei`** (they still **`depends_on`** Neo4j healthy so the stack and **`ingest-worker`** are ordered safely):

```bash
INGEST_MODE=nats docker compose -f docker-compose.yml -f docker-compose.scrape-nats.yml --profile scrape up --build -d
```

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

## Scrape / pack-build profile

```bash
docker compose --profile scrape up --build -d
```

Starts **`proxybroker`**, **`nats`**, **`ingest-worker`**, **`vuln`**, **`lola`**, **`ds`**, **`ti`**, **`sbom`**, **`coderules`**, **`nuclei`**. Configure scrapers to use the broker via `VULN_PROXY_URLS`, `LOLA_PROXY_URLS`, `DS_PROXY_URLS`, `TI_PROXY_URLS` (e.g. `http://proxybroker:8099`) in your override if needed.

After data is in Neo4j, export a pack from the host:

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.0 ./scripts/build-graph-pack.sh
```

### Smoke checklist

1. **Default stack (no scrape):** `docker compose up --build -d` → wait for **`api` healthy** → `curl` **`/health`** and a few **`/v1/...`** calls (see [curl examples](#curl-examples)).
2. **Graph pack without GitHub download:** place **`threat-intel-graph-v0.3.0.zip`** under `data/neo4j_user_export/releases/` and run `docker compose -f docker-compose.yml -f docker-compose.testpack.yml up --build -d` (see [docker-compose.testpack.yml](../docker-compose.testpack.yml)).
3. **Scrape + NATS:** `INGEST_MODE=nats` with **`ingest-worker`** and at least one publisher (e.g. `sbom`); confirm JetStream drains and Neo4j gains expected nodes (sample Cypher per domain in [scrapers/README.md](../scrapers/README.md) / per-module README).
4. **Release asset:** the default URL in [docker/graph-bootstrap.sh](../docker/graph-bootstrap.sh) must point at a ZIP that contains **`manifest.json`** + **`graph.cypher`** with matching **`sha256`**. Bump version and URLs if the dump changes.

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

Replace `vuln` / `Vulnerability` with other [categories](../graph/query/categories.go) and labels as needed.

## MCP (stdio)

Same categorical logic as the API (shared [graph/query](../graph/query) package). Run against a running Neo4j (after bootstrap or scrape).

**Compose (recommended):** start the default stack, then attach MCP on the Compose network (hostname `neo4j`):

```bash
docker compose up --build -d
docker compose --profile mcp run --rm -i mcp
```

**Standalone image** (Neo4j on the host, published Bolt `7687`):

```bash
docker build -f docker/mcp.Dockerfile -t threat_intelligence-mcp .
docker run --rm -i --network host \
  -e NEO4J_URI=neo4j://127.0.0.1:7687 \
  -e NEO4J_USER=neo4j -e NEO4J_PASS=neo4jpassword \
  threat_intelligence-mcp
```

Or from source: `cd mcp && go run ./cmd`.

Category-first tools: `ti_list_categories`, `ti_list_kinds_in_category`, `ti_nodes_by_category`, `ti_search_in_category`; legacy tools remain for raw label access.

## Optional: local pack file

**Quick smoke test with the repo’s sample pack** (no download):

```bash
docker compose -f docker-compose.yml -f docker-compose.testpack.yml up --build -d
```

See [docker-compose.testpack.yml](../docker-compose.testpack.yml) (bind-mounts `data/neo4j_user_export/releases/threat-intel-graph-v0.3.0.zip` as `/pack/host.zip` and sets `GRAPH_PACK_DEFAULT=0`).

Re-importing the same pack into **non-empty** Neo4j data (existing constraints) will fail. For a clean ZIP import use `docker compose … down -v` (drops volumes) or a fresh database.

Create `docker-compose.override.yml` (gitignored by convention or not committed) with:

```yaml
services:
  graph-bootstrap:
    volumes:
      - ./my-pack.zip:/pack/host.zip:ro
```

The bootstrap script copies `/pack/host.zip` when it exists and is non-empty.
