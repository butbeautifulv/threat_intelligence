# Guidance for automated agents (Cursor, CI bots, etc.)

When you change **application code**, **scrapers**, **`ingest-worker`**, **Compose**, or **graph ingest** in this repository:

1. **Read and follow [docs/coding-style.md](docs/coding-style.md)** before writing or reviewing Go: layering (`cmd` → `usecase` → `repository` → `storage`), `log/slog`, lifecycle (`errgroup` + cancel on shutdown for long-running binaries), NATS envelopes in [pkg/ingestv1](pkg/ingestv1).
2. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests to run and docs to update when behaviour or env vars change.
3. Runtime and service matrix: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**; ingest contract (kinds, subjects, ack rules): **[docs/ingest-contract.md](docs/ingest-contract.md)**; scraper matrix: **[scrapers/README.md](scrapers/README.md)**.

Do not invent a parallel style: mirror existing modules cited in `docs/coding-style.md` (for example [scrapers/ti](scrapers/ti), [scrapers/vuln](scrapers/vuln)).
