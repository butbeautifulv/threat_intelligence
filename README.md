# Threat Intelligence

Neo4j-backed graph of vulnerabilities, LOLbins-style artifacts, detection content (Sigma/YARA/Atomic/Caldera), and TI feeds. This repo splits **graph + UX + tooling** (root) from **scrapers / ingest** ([scrapers/README.md](scrapers/README.md)).

## Quick start

```bash
docker compose up --build
```

- **Neo4j Browser:** `http://localhost:7474` — `neo4j` / `neo4jpassword` (defaults in compose). APOC enabled for Cypher export.
- **Panel:** `http://localhost:8088` — force-directed graph, node `markdown` viewer.

Layout:

| Path | Role |
|------|------|
| [graph/](graph/) | Shared Neo4j driver helpers |
| [panel/](panel/) | Next.js graph UI |
| [mcp/](mcp/) | MCP server over the graph |
| [proxybroker/](proxybroker/) | Optional HTTP proxy pool for scrapers |
| [scripts/](scripts/) | Export / pack / import Cypher (`export-graph-cypher.sh`, `build-graph-pack.sh`, `import-graph-pack.sh`) |
| [docs/](docs/) | e.g. [graph pack manifest schema](docs/graph-pack-manifest.schema.json) |
| [scrapers/](scrapers/) | `vuln`, `lola`, `ds`, `ti` ingest binaries + [cue_schemas/](scrapers/cue_schemas/) |

## Offline graph packs (no scraping on target)

After you have filled Neo4j once (see scrapers), ship a versioned ZIP for air-gapped installs:

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v2026.05.0 ./scripts/build-graph-pack.sh
```

Import: [scripts/import-graph-pack.sh](scripts/import-graph-pack.sh) — details in [scrapers/README.md](scrapers/README.md) (section *Graph export and packs*) or inline comments in the script.

## MCP

Build/run from [mcp/](mcp/) (see module `README` if present, or `go run ./cmd` with Neo4j env vars matching compose).

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n) AS labels, count(*) AS c ORDER BY c DESC;
MATCH ()-[r]->() RETURN type(r) AS rel, count(*) AS c ORDER BY c DESC;
```

## Further reading

- **[scrapers/README.md](scrapers/README.md)** — source coverage matrix, env vars, local `go run`, TI JSONL shapes, roadmap.
- **Stage 2:** Kafka workers, STIX/MISP, Cue in CI — sketched in scrapers README.
