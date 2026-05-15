# Guidance for automated agents (Cursor, CI bots, etc.)

## Before you change code

1. **Read and follow [docs/coding-style.md](docs/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; three isolated layers (`scrape/`, `pipeline/`, `graph/`); domain packages per source; shared wire types in `pkg/`. Before merge, check the [PR checklist](docs/coding-style.md#pr-checklist).
2. **Do not add root `go.work`** or cross-layer Go imports between `scrape/`, `pipeline/`, `graph/`. Layers integrate via NATS only; all layers may import `pkg/*`.
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; when changing [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update [docs/schemas/](docs/schemas/) manually in the same PR.
4. Runtime and deploy: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**, **[docs/ingest-contract.md](docs/ingest-contract.md)**, **[deploy/README.md](deploy/README.md)**.
5. Versions: **[versions.env](versions.env)** is the single source of truth for `APP_VERSION` and `GRAPH_PACK_VERSION`.

Reference modules: [scrape/harvest/internal/sources/ti/](scrape/harvest/internal/sources/ti/), [graph/ingest/internal/sources/ti/](graph/ingest/internal/sources/ti/), [pipeline/ned/internal/sources/ti/](pipeline/ned/internal/sources/ti/).

## End-of-task checklist (required)

Complete every step that applies before you consider the task done:

1. **Tests** — run layer targets from repo root: `make test-scrape`, `make test-pipeline`, `make test-graph` for the layers you touched.
2. **Graph version** — if you changed ingest-affecting paths (`scrape/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `graph/ingest/internal/sources/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`), run `./scripts/release/bump-graph-version.sh patch` and rebuild/publish the graph pack when a new ZIP is needed.
3. **Pre-commit check** — `./scripts/release/check-graph-version-bump.sh` (or `make check-graph-version`).
4. **Commit** — descriptive message (what changed and why). Do not commit secrets, `data/`, or `.cursor/plans/`. Use `git add -A -- . ':!data'` when `data/` causes permission errors.
5. **Push** — `git push origin HEAD` unless the user explicitly forbade push or there is no remote.

## Graph pack releases

- Default version: see [versions.env](versions.env).
- Workflow: [docs/graph-pack.md](docs/graph-pack.md).
- Publish: `GRAPH_PACK_VERSION=vX.Y.Z ./scripts/release/publish-graph-pack.sh`.
