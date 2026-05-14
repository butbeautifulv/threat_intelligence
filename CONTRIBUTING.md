# Contributing

Thank you for improving this project. Small, focused changes are easier to review than large mixed diffs.

## Before you open a PR

1. Read **[docs/coding-style.md](docs/coding-style.md)** — layers (`cmd` → `usecase` → `repository` → `storage` / `feeds`), `slog`, and ingest conventions. **Automated agents:** follow **[AGENTS.md](AGENTS.md)** so changes stay consistent with this doc.
2. Follow **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** in issues, reviews, and discussions.
3. Run tests for the modules you touched, for example:
   - `go test ./pkg/ingestv1/... ./scrapers/ingestpub/... ./scrapers/sbom/... ./scrapers/coderules/... ./scrapers/nuclei/... ./scrapers/ingest-worker/... ./scrapers/ti/... ./scrapers/vuln/... ./scrapers/lola/... ./scrapers/ds/... ./api/... ./graph/... ./mcp/...`
   - or from a single scraper directory: `go test ./...`
4. **Read layer sanity (API / MCP):** they must not import NATS or ingest packages. From repo root, expect **no output** (and `grep` exit status **1**):
   - `grep -R --include='*.go' -E 'ingestv1|ingestpub|nats\\.io|NATS_' api mcp`
5. If you change Compose behaviour or env vars, update **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)** and **[scrapers/README.md](scrapers/README.md)** in the same PR.

## Commits and branches

- Prefer descriptive commit messages (what changed and why).
- Rebase or merge with your team’s usual Git workflow; avoid force-pushing shared branches without agreement.

## Licensing

By contributing, you agree that your contributions are licensed under the **MIT License** ([LICENSE](LICENSE)), the same as the rest of the repository. See [SECURITY.md](SECURITY.md) for responsible disclosure.
