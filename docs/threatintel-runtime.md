# Threat Intel runtime (Docker Compose)

Default stack: **Neo4j** → **graph-bootstrap** (import graph pack) → **HTTP API**. Live scraping, **NATS**, **ingest-worker**, and **proxybroker** are opt-in via the **`scrape`** profile. **MCP** uses **`profiles: ["mcp"]`**.

## Ports

| Service | Port | Notes |
|---------|------|--------|
| Neo4j Browser | `${NEO4J_HTTP_PORT:-7474}` (host) | Bolt `${NEO4J_BOLT_PORT:-7687}`; map with `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` if defaults are busy |
| HTTP API | 8090 | `API_PORT` to override published port |
| Proxybroker | 8099 | Only with `--profile scrape`; `PROXYBROKER_PORT` |
| NATS client | `${NATS_CLIENT_PORT:-4222}` | Only with `--profile scrape`; maps container `4222` |
| NATS monitoring | `${NATS_MONITOR_PORT:-8222}` | HTTP monitoring on container port `8222` (e.g. `http://localhost:8222` when published) |

## Compose service reference

All definitions live in [docker-compose.yml](../docker-compose.yml). **Binary / module docs:** see links in each subsection.

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
| **Purpose** | Message bus for optional **`INGEST_MODE=nats`** on `sbom`, `coderules`, `nuclei` |
| **Ports** | Client and monitoring (see [Ports](#ports)) |
| **Stream** | Created by publishers or **`ingest-worker`**: name **`INGEST`**, subjects **`ingest.appsec.>`** |

### ingest-worker

| | |
|--|--|
| **Profile** | `scrape` |
| **Build** | [docker/ingest-worker.Dockerfile](../docker/ingest-worker.Dockerfile) |
| **Module** | [scrapers/ingest-worker/README.md](../scrapers/ingest-worker/README.md) |
| **Purpose** | Long-running **JetStream pull consumer**: reads `ingestv1` envelopes from **`ingest.appsec.>`**, writes **Neo4j** using the same `MERGE` paths as **`sbom`**, **`coderules`**, and **`nuclei`** in `INGEST_MODE=direct` |
| **Depends on** | `neo4j` healthy, `nats` started |
| **Restart** | `on-failure` |
| **Env** | `NEO4J_*`, `NATS_URL` (Compose: `nats://nats:4222`), `NATS_INGEST_STREAM`, `NATS_DURABLE`, `NATS_SUBSCRIBE_SUBJECT`, `INGEST_BATCH`, `INGEST_MAX_WAIT` — full table in [scrapers/ingest-worker/README.md](../scrapers/ingest-worker/README.md) |

Use **`ingest-worker`** whenever AppSec scrapers publish with **`INGEST_MODE=nats`**; otherwise messages stay in JetStream until drained.

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
| `vuln` | [docker/vuln.Dockerfile](../docker/vuln.Dockerfile) | NVD, Metasploit, Exploit-DB, optional Vulners; volume `data/cache` |
| `lola` | [docker/lola.Dockerfile](../docker/lola.Dockerfile) | LOLBAS, GTFOBins, LOFTS, MITRE STIX |
| `ds` | [docker/ds.Dockerfile](../docker/ds.Dockerfile) | Sigma, YARA, Atomic, Caldera |
| `ti` | [docker/ti.Dockerfile](../docker/ti.Dockerfile) | KEV, URLhaus, ThreatFox, …; `TI_FEEDS`, `TI_*` limits |
| `sbom` | [docker/sbom.Dockerfile](../docker/sbom.Dockerfile) | OSV + GHSA; **`INGEST_MODE`**, **`SBOM_NATS_SUBJECT`**; depends on `nats` (started) |
| `coderules` | [docker/coderules.Dockerfile](../docker/coderules.Dockerfile) | CWE, Semgrep, CodeQL; **`INGEST_MODE`**, **`CODERULES_NATS_SUBJECT`**; depends on `nats` |
| `nuclei` | [docker/nuclei.Dockerfile](../docker/nuclei.Dockerfile) | Nuclei templates; **`INGEST_MODE`**, **`NUCLEI_NATS_SUBJECT`**; depends on `nats` |

All scrape ingest rows above use **`NEO4J_URI=neo4j://neo4j:7687`** (except **`coderules` / `nuclei` / `sbom`** do not open Neo4j when **`INGEST_MODE=nats`**; **`sbom`** uses **`SBOM_CVE_LIST_FILE`** or **`SBOM_CVE_LIST_URL`** for OSV CVE ids in `nats` mode). See [scrapers/README.md](../scrapers/README.md) for per-feed env vars.

### NATS publish and consume (`INGEST_MODE`)

| Variable | Default | Meaning |
|----------|---------|--------|
| `INGEST_MODE` | `direct` | `direct` = scraper writes Neo4j; `nats` = publish envelopes (AppSec scrapers only, as wired in compose) |
| `NATS_URL` | `nats://nats:4222` in Compose | NATS client URL for publishers and **ingest-worker** |
| `SBOM_NATS_SUBJECT` | `ingest.appsec.sbom` | Publish subject for `sbom` |
| `SBOM_CVE_LIST_FILE` | empty (Compose scrape sets image default) | CVE list for OSV when **`INGEST_MODE=nats`** |
| `SBOM_CVE_LIST_URL` | empty | Alternative CVE list URL if file unset |
| `CODERULES_NATS_SUBJECT` | `ingest.appsec.coderules` | Publish subject for `coderules` |
| `NUCLEI_NATS_SUBJECT` | `ingest.appsec.nuclei` | Publish subject for `nuclei` |
| `NATS_INGEST_STREAM` | `INGEST` | Stream name (worker) |
| `NATS_DURABLE` | `ingest-worker` | Durable consumer name |
| `NATS_SUBSCRIBE_SUBJECT` | `ingest.appsec.>` | Worker pull filter |
| `INGEST_BATCH` | `10` | Max messages per fetch |
| `INGEST_MAX_WAIT` | `5s` | Fetch wait |

JetStream dedup: **`Nats-Msg-Id`** from envelope **`idempotency_key`** ([scrapers/ingestpub](../scrapers/ingestpub), [pkg/ingestv1](../pkg/ingestv1)).

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
