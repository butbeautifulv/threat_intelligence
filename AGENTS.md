# Guidance for automated agents (Cursor, CI bots, etc.)

When you change application code, deploy, or contracts in this repository:

1. **Read and follow [docs/coding-style.md](docs/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; three isolated layers (`scrape/`, `pipeline/`, `graph/`); mandatory `internal/domain/`; schema-first contracts.
2. **Do not add root `go.work`** or cross-layer Go imports. Layers integrate via NATS only.
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; run `scripts/gen-contracts.sh` when [docs/schemas/](docs/schemas/) change.
4. Runtime and deploy: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**, **[docs/ingest-contract.md](docs/ingest-contract.md)**, **[deploy/](deploy/)**.

Reference modules: [scrape/sources/ti](scrape/sources/ti), [graph/sources/ti](graph/sources/ti), [pipeline/pipeline_worker](pipeline/pipeline_worker).
