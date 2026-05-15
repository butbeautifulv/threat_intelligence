# Scripts

Host-side helpers for **Neo4j export**, **graph pack** releases, **stack ops**, **E2E smoke**, and **housekeeping**. Not the pipeline NED runtime ([pipeline/ned](../pipeline/ned/)).

Shared library: [lib/common.sh](lib/common.sh) (`COMPOSE_FILES`, `compose()`, pack naming `veil-graph-v*`).

## Layout

| Path | Purpose |
|------|---------|
| [lib/common.sh](lib/common.sh) | Compose helpers, `veil-graph-v*` pack names, profiles |
| [ops/compose-up-full.sh](ops/compose-up-full.sh) | Full stack up (optional worker scale / scrape partition) |
| [graph-pack/export-cypher.sh](graph-pack/export-cypher.sh) | APOC export → `data/neo4j_user_export/graph.cypher` |
| [graph-pack/build.sh](graph-pack/build.sh) | ZIP + `manifest.json` → `data/.../releases/veil-graph-vX.zip` |
| [graph-pack/import.sh](graph-pack/import.sh) | Import pack (supports legacy `threat-intel-graph-*.zip`) |
| [graph-pack/profile-fast-rich.sh](graph-pack/profile-fast-rich.sh) | ~25 min crawl profile → compose-up-full |
| [release/publish-graph-pack.sh](release/publish-graph-pack.sh) | Build + `gh release create veil-graph-vX` |
| [test/smoke-scrape-e2e.sh](test/smoke-scrape-e2e.sh) | E2E smoke (default profile [deploy/profiles/smoke-minimal.env](../deploy/profiles/smoke-minimal.env)) |
| [test/verify-nvd-enrichment.sh](test/verify-nvd-enrichment.sh) | Cypher QA for NVD CWE/CPE |
| [housekeeping/graph-dedup-cleanup.sh](housekeeping/graph-dedup-cleanup.sh) | Post-ingest Neo4j dedup |

Deploy profiles: [deploy/profiles/](../deploy/profiles/). Runtime: [docs/threatintel-runtime.md](../docs/threatintel-runtime.md). Graph pack workflow: [docs/graph-pack.md](../docs/graph-pack.md).

## Graph pack workflow

```bash
./scripts/graph-pack/profile-fast-rich.sh    # optional: full crawl
./scripts/graph-pack/export-cypher.sh
GRAPH_PACK_VERSION=v0.4.0 ./scripts/graph-pack/build.sh
GRAPH_PACK_VERSION=v0.4.0 ./scripts/release/publish-graph-pack.sh --skip-build
```

Import locally:

```bash
USE_DOCKER_COMPOSE=1 ./scripts/graph-pack/import.sh \
  data/neo4j_user_export/releases/veil-graph-v0.4.0.zip
```

## Quick commands

```bash
./scripts/ops/compose-up-full.sh
./scripts/test/smoke-scrape-e2e.sh --up
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/ops/compose-up-full.sh
```
