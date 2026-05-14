# Threat Intel runtime (Docker Compose)

Default stack: **Neo4j** → **graph-bootstrap** (import graph pack) → **HTTP API**. Live scraping (including **`proxybroker`**) is opt-in via the **`scrape`** profile only.

## Ports

| Service | Port | Notes |
|---------|------|--------|
| Neo4j Browser | `${NEO4J_HTTP_PORT:-7474}` (host) | Bolt `${NEO4J_BOLT_PORT:-7687}`; map with `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` if defaults are busy |
| HTTP API | 8090 | `API_PORT` to override published port |
| Proxybroker | 8099 | Only with `--profile scrape`; `PROXYBROKER_PORT` |

## Graph bootstrap (usage mode)

Init container `graph-bootstrap` runs once after Neo4j is healthy. Pack resolution order:

1. Bind mount **`/pack/host.zip`** inside the container (optional; add under `graph-bootstrap.volumes` in a local override file).
2. **`GRAPH_PACK_FILE`** — path *inside the container* to a `.zip` (if you extend the service with a volume).
3. **`GRAPH_PACK_URL`** — HTTP(S) URL to a pack ZIP.
4. If **`GRAPH_PACK_DEFAULT=1`** (default): download **`GRAPH_PACK_DEFAULT_URL`**, or the built-in default release asset if unset.

Skip import entirely: **`GRAPH_PACK_SKIP=1`**.

| Variable | Default | Meaning |
|----------|---------|---------|
| `GRAPH_PACK_SKIP` | `0` | `1` = exit 0 without importing |
| `GRAPH_PACK_DEFAULT` | `1` | `0` = do not download the default release ZIP when no file/URL |
| `GRAPH_PACK_URL` | empty | HTTP(S) URL of the pack ZIP |
| `GRAPH_PACK_DEFAULT_URL` | built-in GitHub asset | Overrides the default download URL when `GRAPH_PACK_DEFAULT=1` |
| `GRAPH_PACK_FILE` | empty | Path **inside** the bootstrap container (mount a volume if needed) |

Compose passes these from the host for `graph-bootstrap` (see [docker-compose.yml](../docker-compose.yml) `environment`).

Checksum: `manifest.json` `sha256` must match `graph.cypher` (same rules as [scripts/import-graph-pack.sh](../scripts/import-graph-pack.sh)).

## Scrape / pack-build profile

Build and run scrapers + proxy pool:

```bash
docker compose --profile scrape up --build -d
```

Services (all `profiles: ["scrape"]`): `proxybroker`, `vuln`, `lola`, `ds`, `ti`. Configure scrapers to use the broker via `VULN_PROXY_URLS`, `LOLA_PROXY_URLS`, `DS_PROXY_URLS`, `TI_PROXY_URLS` (e.g. `http://proxybroker:8099`) in your override if needed.

After data is in Neo4j, export a pack from the host:

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v2026.05.0 ./scripts/build-graph-pack.sh
```

## HTTP API (categorical)

Base URL: `http://localhost:8090` (with default compose).

- `GET /health`
- `GET /v1/categories` — product categories (`vuln`, `ti`, `detection`, `lola`, `mitre`) and Neo4j label sets.
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

See [docker-compose.testpack.yml](../docker-compose.testpack.yml) (bind-mounts `data/neo4j_user_export/releases/threat-intel-graph-v0.2.0.zip` as `/pack/host.zip` and sets `GRAPH_PACK_DEFAULT=0`).

Re-importing the same pack into **non-empty** Neo4j data (existing constraints) will fail. For a clean ZIP import use `docker compose … down -v` (drops volumes) or a fresh database.

Create `docker-compose.override.yml` (gitignored by convention or not committed) with:

```yaml
services:
  graph-bootstrap:
    volumes:
      - ./my-pack.zip:/pack/host.zip:ro
```

The bootstrap script copies `/pack/host.zip` when it exists and is non-empty.
