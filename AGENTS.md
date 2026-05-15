# Guidance for automated agents (Cursor, CI bots, etc.)

When you change application code, deploy, or envelope types in this repository:

1. **Read and follow [docs/coding-style.md](docs/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; three isolated layers (`scrape/`, `pipeline/`, `graph/`); domain packages per source; shared wire types in `pkg/`. Before merge, check the [PR checklist](docs/coding-style.md#pr-checklist).
2. **Do not add root `go.work`** or cross-layer Go imports between `scrape/`, `pipeline/`, `graph/`. Layers integrate via NATS only; all layers may import `pkg/*`.
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; when changing [pkg/harvest](pkg/harvest/) or [pkg/commit](pkg/commit/), update [docs/schemas/](docs/schemas/) manually in the same PR.
4. Runtime and deploy: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**, **[docs/ingest-contract.md](docs/ingest-contract.md)**, **[deploy/README.md](deploy/README.md)**.

Reference modules: [scrape/harvest/internal/sources/ti](scrape/harvest/internal/sources/ti), [graph/ingest/internal/sources/ti](graph/ingest/internal/sources/ti), [pipeline/ned/internal/sources/ti](pipeline/ned/internal/sources/ti).
