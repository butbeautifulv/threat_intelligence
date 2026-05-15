# Graph pack (export, release, import)

Versioned Neo4j snapshots as **`veil-graph-vX.Y.Z.zip`** + `manifest.json`. GitHub release tag: **`veil-graph-vX.Y.Z`** (same as ZIP basename without `.zip`).

Current default version: **[versions.env](../versions.env)** (`GRAPH_PACK_VERSION`).

## Artifact layout

| File | Role |
|------|------|
| `manifest.json` | `schema` (`veil.graph-pack/1`), `graphVersion`, `sha256`, `createdAt` |
| `graph.cypher` | APOC `apoc.export.cypher.all` dump |

Legacy packs (`threat-intel-graph-*.zip`, schema `threat-intelligence.graph-pack/1`) still import via [scripts/graph-pack/import.sh](../scripts/graph-pack/import.sh).

## Build (local)

```bash
./scripts/graph-pack/export-cypher.sh
./scripts/graph-pack/build.sh
```

Output: `data/neo4j_user_export/releases/veil-graph-vX.Y.Z.zip` (version from `versions.env` or `GRAPH_PACK_VERSION` env).

Fast-rich crawl profile (~25 min): [scripts/graph-pack/profile-fast-rich.sh](../scripts/graph-pack/profile-fast-rich.sh) ([deploy/profiles/fast-rich.env](../deploy/profiles/fast-rich.env)).

## Publish (GitHub)

```bash
./scripts/release/publish-graph-pack.sh
# or after manual build:
./scripts/release/publish-graph-pack.sh --skip-build
```

After ingest-affecting code changes, bump first: `./scripts/release/bump-graph-version.sh patch`.

Sets `GRAPH_PACK_DEFAULT_URL` for [graph-bootstrap](../deploy/graph/docker/graph-bootstrap.sh) (`GRAPH_PACK_DEFAULT_VERSION` from `versions.env`).

## Import

- **Compose bootstrap:** `graph-bootstrap` service (default download or `GRAPH_PACK_SKIP=1`)
- **Host:** `USE_DOCKER_COMPOSE=1 ./scripts/graph-pack/import.sh <url-or-zip>`
- **Local ZIP test:** [docker-compose.testpack.yml](../docker-compose.testpack.yml)

Manifest schema: [graph-pack-manifest.schema.json](graph-pack-manifest.schema.json).

## Related

- [scripts/README.md](../scripts/README.md) — script index
- [deploy/README.md](../deploy/README.md) — scaling, smoke
- [threatintel-runtime.md](threatintel-runtime.md) — compose services and env
