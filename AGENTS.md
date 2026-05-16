# Guidance for automated agents (Cursor, CI bots, etc.)

## Before you change code

1. **Read and follow [docs/coding-style.md](docs/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; four isolated contexts (`scrape/`, `pipeline/`, `graph/`, `engage/`); domain packages per source; shared wire types in `pkg/`. Before merge, check the [PR checklist](docs/coding-style.md#pr-checklist).
2. **Do not add root `go.work`** or cross-layer Go imports between `scrape/`, `pipeline/`, `graph/`, `engage/`. Scrape/pipeline/graph integrate via NATS; engage calls graph only via HTTP veil-api; all layers may import `pkg/*`.
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; when changing [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update [docs/schemas/](docs/schemas/) manually in the same PR.
4. Runtime and deploy: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**, **[docs/ingest-contract.md](docs/ingest-contract.md)**, **[deploy/README.md](deploy/README.md)**.
5. Versions: **[versions.env](versions.env)** is the single source of truth for `APP_VERSION` and `GRAPH_PACK_VERSION`.

Reference modules: [scrape/harvest/internal/sources/ti/](scrape/harvest/internal/sources/ti/), [graph/ingest/internal/sources/ti/](graph/ingest/internal/sources/ti/), [pipeline/ned/internal/sources/ti/](pipeline/ned/internal/sources/ti/).

## Planning and commit rhythm (required for multi-phase work)

Keep diffs reviewable: **one git commit per completed phase or slice**, not one giant commit at the end.

1. **Master plan** — before coding, write or update a master plan (status table, phase list, dependencies). For Engage work, keep it under `.cursor/plans/` (e.g. `engage_hexstrike_master_*.plan.md`, `engage/engage_phase_*.plan.md`).
2. **Phase plan** — for the active phase only, add or open a slice plan derived from the master plan (scope, files, acceptance).
3. **Execute one phase** — implement only what that phase plan covers; run tests for touched layers.
4. **Commit the phase** — `git add` + commit with a message like `feat(engage): Phase N — <short title>`. Do not start the next phase with a dirty tree unless the user asked to batch.
5. **Update master plan** — mark the phase done, note commit SHA or branch state, adjust follow-ups.
6. **Push** — after each phase commit (or when the user asks), `git push origin HEAD`.

If the user asks to “stage all” or catch up after many phases, still document the phase boundaries in the commit message body.

## End-of-task checklist (required)

Complete every step that applies before you consider the task done:

1. **Tests** — run layer targets from repo root: `make test-scrape`, `make test-pipeline`, `make test-graph`, `make test-engage` for the layers you touched. For `graph/serve` only: `make test-graph-serve` (`-race`). Graph read Docker smoke: `make test-graph-read-smoke`. Engage: `make test-engage-parity` when changing catalog. Engage events bus (`engage/.../events`, `pipeline/engage-events`, `graph/ingest/.../engage`): also `make test-pipeline`; optional Docker `make test-engage-events-pipeline`.
2. **Graph version** — if you changed ingest-affecting paths (`scrape/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `graph/ingest/internal/sources/` including `engage/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`), run `./scripts/release/bump-graph-version.sh patch` and rebuild/publish the graph pack when a new ZIP is needed.
3. **Pre-commit check** — `./scripts/release/check-graph-version-bump.sh` (or `make check-graph-version`).
4. **Commit** — descriptive message (what changed and why). Do not commit secrets or `data/`. Use `git add -A -- . ':!data'` when `data/` causes permission errors. Exclude `**/__pycache__/`.
5. **Push** — `git push origin HEAD` unless the user explicitly forbade push or there is no remote.

## Graph pack releases

- Default version: see [versions.env](versions.env).
- Workflow: [docs/graph-pack.md](docs/graph-pack.md).
- Publish: `GRAPH_PACK_VERSION=vX.Y.Z ./scripts/release/publish-graph-pack.sh`.
