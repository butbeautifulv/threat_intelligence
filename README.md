# Threat Intelligence

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Neo4j-backed graph of vulnerabilities, LOLbins-style artifacts, detection content (Sigma/YARA/Atomic/Caldera), and TI feeds. The repository splits **graph + HTTP API + MCP** (root modules) from **scrapers and ingest** ([scrapers/README.md](scrapers/README.md)).

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Security:** [SECURITY.md](SECURITY.md) · **Code of conduct:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Quick start

```bash
docker compose up --build
```

Default stack: **Neo4j** → **graph-bootstrap** (imports a [graph pack](docs/threatintel-runtime.md#graph-bootstrap-usage-mode); skip with `GRAPH_PACK_SKIP=1`) → **HTTP API**. Scrapers, NATS, and **ingest-worker** are **not** started unless you use the **`scrape`** profile.

| Endpoint | Default | Notes |
|----------|---------|--------|
| Neo4j Browser | `http://localhost:7474` | User `neo4j` / `neo4jpassword`; APOC enabled |
| HTTP API | `http://localhost:8090` | `API_PORT` overrides published port |

**Fill the graph by scraping** (includes optional **NATS** + **ingest-worker** for queue-backed AppSec ingest):

```bash
docker compose --profile scrape up --build -d
```

Full service matrix, ports, and environment variables: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**.

## Documentation index

| Document | Contents |
|----------|----------|
| [docs/threatintel-runtime.md](docs/threatintel-runtime.md) | Compose services (including **ingest-worker**), ports, bootstrap, API, MCP, NATS / `INGEST_MODE` |
| [scrapers/README.md](scrapers/README.md) | Scraper sources matrix, env vars, `direct` vs `nats`, local `go run` |
| [scrapers/ingest-worker/README.md](scrapers/ingest-worker/README.md) | JetStream consumer: env, local run, Compose examples |
| [docs/coding-style.md](docs/coding-style.md) | Layering, logging, ingest conventions for PRs |
| [docs/ontology-appsec.md](docs/ontology-appsec.md) | AppSec labels, relationships, roadmap |
| [mcp/README.md](mcp/README.md) | Stdio MCP server tools and env |
| [scripts/README.md](scripts/README.md) | Export, packs, graph housekeeping |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute, tests, licensing |
| [SECURITY.md](SECURITY.md) | Responsible disclosure for vulnerabilities |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community expectations |

## Repository layout

| Path | Role |
|------|------|
| [graph/](graph/) | Shared Neo4j client + [graph/query](graph/query) (categories, Cypher reads) |
| [api/](api/) | HTTP API (`docker/api.Dockerfile`) |
| [docker/](docker/) | Service Dockerfiles: `api`, `graph-bootstrap`, scrapers, `mcp`, **`ingest-worker`**, `proxybroker`, … |
| [mcp/](mcp/) | MCP server (stdio); uses `graph/query` |
| [pkg/ingestv1/](pkg/ingestv1/) | Versioned JSON envelope for NATS → worker pipeline |
| [scrapers/](scrapers/) | `vuln`, `lola`, `ds`, `ti`, `sbom`, `coderules`, `nuclei`, **`ingest-worker`**, `ingestpub`, `proxybroker`, [cue_schemas/](scrapers/cue_schemas/) |
| [scripts/](scripts/) | Export / pack / import Cypher; [scripts/README.md](scripts/README.md) |
| [docs/](docs/) | Runtime, ontology, OpenAPI sketch, coding style |

## Offline graph packs

After Neo4j has been filled once, ship a versioned ZIP for air-gapped installs. Exports go under `data/neo4j_user_export/` (see `scripts/export-graph-cypher.sh`).

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v2026.05.0 ./scripts/build-graph-pack.sh
```

Import: [scripts/import-graph-pack.sh](scripts/import-graph-pack.sh) — details in [scrapers/README.md](scrapers/README.md) (*Graph export and packs*).

## MCP

```bash
docker compose --profile mcp run --rm -i mcp
```

Details: [docs/threatintel-runtime.md](docs/threatintel-runtime.md#mcp-stdio) and [mcp/README.md](mcp/README.md).

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n) AS labels, count(*) AS c ORDER BY c DESC;
MATCH ()-[r]->() RETURN type(r) AS rel, count(*) AS c ORDER BY c DESC;
```

## Further reading

- **[docs/coding-style.md](docs/coding-style.md)** — scraper and worker layering, `slog`, optional NATS ingest.
- **[scrapers/README.md](scrapers/README.md)** — source matrix, `INGEST_MODE`, TI JSONL, roadmap.
- **Stage 2:** Kafka workers, STIX/MISP, Cue in CI — sketched in scrapers README.
