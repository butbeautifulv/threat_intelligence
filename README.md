# Veil (Vulnerability Exploitation Intelligence Layer)

![Veil](docs/veil.png)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Veil** is a Neo4j-backed threat-intelligence graph: vulnerabilities (CVE, CWE, CPE), LOLbins-style artifacts, detection content (Sigma/YARA/Caldera), TI feeds, SBOM advisories, and code-rule templates. The runtime is split into three isolated layers — **scrape**, **pipeline**, and **graph** — connected by NATS JetStream (`scrape.>` → `ingest.>`).

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Agents / AI:** [AGENTS.md](AGENTS.md) · **Security:** [SECURITY.md](SECURITY.md) · **Code of conduct:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Architecture

```mermaid
flowchart LR
  subgraph scrape [scrape]
    SW[scrape_worker]
    SW -->|scrapev1| NATS1[NATS SCRAPE]
  end
  subgraph pipeline [pipeline]
    PW[pipeline_worker]
    NATS1 --> PW
    PW -->|ingestv1| NATS2[NATS INGEST]
  end
  subgraph graph [graph]
    IW[ingest_worker]
    Neo4j[(Neo4j)]
    API[HTTP API]
    NATS2 --> IW --> Neo4j
    API --> Neo4j
  end
```

| Layer | Path | Role |
|-------|------|------|
| **Scrape** | [scrape/](scrape/) | Fetch feeds, Vitess ledger, publish `scrapev1` |
| **Pipeline** | [pipeline/](pipeline/) | Normalize/dedup → `ingestv1` (incl. NVD CWE/CPE via [pkg/nvdparse](pkg/nvdparse/)) |
| **Graph** | [graph/](graph/) | MERGE into Neo4j; [api/](graph/api/), [mcp/](graph/mcp/) read Bolt |

Deploy: [deploy/](deploy/) · Contracts: [docs/ingest-contract.md](docs/ingest-contract.md) · Runtime: [docs/threatintel-runtime.md](docs/threatintel-runtime.md)

## Quick start

### Graph only (demo API + optional pack import)

```bash
docker compose up --build -d
```

| Endpoint | Default |
|----------|---------|
| Neo4j Browser | http://localhost:7474 (`neo4j` / `neo4jpassword`) |
| HTTP API | http://localhost:8090 |

`graph-bootstrap` imports the default [graph pack v0.3.2](https://github.com/butbeautifulv/threat_intelligence/releases/tag/v0.3.2-graph-pack) unless `GRAPH_PACK_SKIP=1`. Local ZIP: [docker-compose.testpack.yml](docker-compose.testpack.yml).

```bash
curl -sS http://localhost:8090/health
curl -sS http://localhost:8090/v1/categories | jq .
```

### Full scrape pipeline

```bash
./scripts/compose-up-full.sh
# or:
# docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml up --build
```

E2E smoke (minimal sources by default):

```bash
./scripts/smoke_scrape_e2e.sh --up
./scripts/smoke_scrape_e2e.sh
```

Fast-rich graph pack build (~25 min): [scripts/graph-pack-run-v032.sh](scripts/graph-pack-run-v032.sh) — see [docs/threatintel-runtime.md](docs/threatintel-runtime.md#fast-rich-graph-pack-profile-25-min).

## Documentation index

| Document | Contents |
|----------|----------|
| [AGENTS.md](AGENTS.md) | Cursor/agents: read [docs/coding-style.md](docs/coding-style.md) first |
| [docs/threatintel-runtime.md](docs/threatintel-runtime.md) | Compose, ports, env, bootstrap, API, MCP, NATS, graph packs |
| [docs/deploy.md](docs/deploy.md) | Worker scaling, full stack, releases |
| [deploy/README.md](deploy/README.md) | Per-layer compose files |
| [scrape/README.md](scrape/README.md) | Scrape sources and env vars |
| [pipeline/README.md](pipeline/README.md) | Pipeline worker and normalization |
| [graph/README.md](graph/README.md) | Ingest, API, MCP, Neo4j client |
| [graph/ingest_worker/README.md](graph/ingest_worker/README.md) | JetStream → Neo4j consumer |
| [docs/coding-style.md](docs/coding-style.md) | Three layers, DDD, contracts |
| [docs/ontology-appsec.md](docs/ontology-appsec.md) | Labels, relationships, roadmap |
| [docs/ingest-contract.md](docs/ingest-contract.md) | `scrapev1` / `ingestv1`, JetStream |
| [graph/mcp/README.md](graph/mcp/README.md) | Stdio MCP tools |
| [scripts/README.md](scripts/README.md) | Export, packs, smoke, dedup |

## Repository layout

| Path | Role |
|------|------|
| [scrape/](scrape/) | `scrape_worker`, [sources/](scrape/sources/), [feeds/](scrape/feeds/), [ledger/](scrape/ledger/), [contract/scrapev1](scrape/contract/scrapev1/) |
| [pipeline/](pipeline/) | `pipeline_worker`, [internal/normalize/](pipeline/internal/normalize/), [contract/ingestv1](pipeline/contract/ingestv1/) |
| [graph/](graph/) | `ingest_worker`, [sources/](graph/sources/), [storage/](graph/storage/), [neo4jclient/](graph/neo4jclient/), [api/](graph/api/), [mcp/](graph/mcp/) |
| [pkg/nvdparse/](pkg/nvdparse/) | Shared NVD JSON parser (CWE, CPE, CVSS) for scrape + pipeline |
| [deploy/](deploy/) | Per-layer Docker Compose and Dockerfiles |
| [scripts/](scripts/) | `compose-up-full.sh`, smoke, export/pack, dedup |
| [docs/](docs/) | Runtime, ontology, JSON schemas |

## Graph packs

After Neo4j has data from scrape or a prior import:

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.2 ./scripts/build-graph-pack.sh
```

Import: [scripts/import-graph-pack.sh](scripts/import-graph-pack.sh). Verify NVD enrichment: [scripts/verify-nvd-enrichment.sh](scripts/verify-nvd-enrichment.sh).

**Note:** Pack **v0.3.2** on GitHub was built before the pipeline enrichment fix; rebuild with current `main` for full `HAS_CWE` / `AFFECTS` / `CPE` in exports.

## MCP

```bash
cd graph/mcp && go run ./cmd
```

Details: [graph/mcp/README.md](graph/mcp/README.md), [docs/threatintel-runtime.md](docs/threatintel-runtime.md#mcp-stdio).

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n)[0] AS label, count(*) AS c ORDER BY c DESC LIMIT 20;
MATCH (v:Vulnerability)-[:HAS_CWE]->() RETURN count(*) AS has_cwe;
MATCH (v:Vulnerability)-[:AFFECTS]->(:CPE) RETURN count(*) AS affects;
```

## Tests

```bash
make contracts    # sync graph/contract/ingestv1 from pipeline
make test-scrape
make test-pipeline
make test-graph
```
