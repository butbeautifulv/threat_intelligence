# Threat Intelligence

Neo4j-backed graph of vulnerabilities, LOLbins-style artifacts, detection content (Sigma/YARA/Atomic/Caldera), and TI feeds. This repo splits **graph + tooling** (root) from **scrapers / ingest** ([scrapers/README.md](scrapers/README.md)).

## Quick start

```bash
docker compose up --build
```

Default behaviour: **Neo4j** starts, **graph-bootstrap** imports a [graph pack](docs/threatintel-runtime.md#graph-bootstrap-usage-mode) (from `GRAPH_PACK_URL` / default GitHub release, or skip with `GRAPH_PACK_SKIP=1`), then **HTTP API** starts. Scrapers are **not** started unless you use the `scrape` profile.

- **Neo4j Browser:** `http://localhost:7474` — `neo4j` / `neo4jpassword` (defaults in compose). APOC enabled for Cypher export.
- **HTTP API:** `http://localhost:8090` — categorical REST (`/v1/categories`, …). See [docs/threatintel-runtime.md](docs/threatintel-runtime.md).

**Fill graph by scraping instead of importing a pack:**

```bash
docker compose --profile scrape up --build -d
```

Runtime details, env vars, MCP over stdio, and OpenAPI sketch: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**.

Layout:

| Path | Role |
|------|------|
| [graph/](graph/) | Shared Neo4j client + [graph/query](graph/query) (categories, Cypher reads) |
| [api/](api/) | HTTP API (`docker/api.Dockerfile`): `cmd`, `internal/config`, `internal/domain`, `internal/storage/neo4j`, `internal/usecase`, `internal/transport/httpserver`, `internal/components` |
| [docker/](docker/) | **All** service Dockerfiles (api, bootstrap, scrapers, mcp, `scrapers/proxybroker`) |
| [mcp/](mcp/) | MCP server (stdio); uses `graph/query` |
| [scrapers/proxybroker/](scrapers/proxybroker/) | HTTP proxy pool for scrapers (compose profile `scrape`) |
| [scripts/](scripts/) | Export / pack / import Cypher; [scripts/README.md](scripts/README.md) (`graph-dedup-cleanup.sh`, …) |
| [docs/](docs/) | [threatintel-runtime.md](docs/threatintel-runtime.md), [ontology-appsec.md](docs/ontology-appsec.md) (labels + roadmap), [openapi.yaml](docs/openapi.yaml), [graph pack manifest schema](docs/graph-pack-manifest.schema.json) |
| [scrapers/](scrapers/) | `vuln`, `lola`, `ds`, `ti`, `sbom`, `coderules`, `nuclei` ingest binaries + [cue_schemas/](scrapers/cue_schemas/) |

## Offline graph packs (no scraping on target)

After you have filled Neo4j once (see scrapers), ship a versioned ZIP for air-gapped installs. Exports land under `data/neo4j_user_export/` (see `scripts/export-graph-cypher.sh`).

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v2026.05.0 ./scripts/build-graph-pack.sh
```

Import: [scripts/import-graph-pack.sh](scripts/import-graph-pack.sh) — details in [scrapers/README.md](scrapers/README.md) (section *Graph export and packs*) or inline comments in the script.

## MCP

Stdio MCP server in [mcp/](mcp/) (same categorical queries as the HTTP API). After `docker compose up`, run:

`docker compose --profile mcp run --rm -i mcp`

Details: [docs/threatintel-runtime.md](docs/threatintel-runtime.md#mcp-stdio) and [mcp/README.md](mcp/README.md).

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n) AS labels, count(*) AS c ORDER BY c DESC;
MATCH ()-[r]->() RETURN type(r) AS rel, count(*) AS c ORDER BY c DESC;
```

## Further reading

- **[scrapers/README.md](scrapers/README.md)** — source coverage matrix, env vars, local `go run`, TI JSONL shapes, roadmap.
- **Stage 2:** Kafka workers, STIX/MISP, Cue in CI — sketched in scrapers README.
