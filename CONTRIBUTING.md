# Contributing

Thank you for improving this project. Small, focused changes are easier to review than large mixed diffs.

## Before you open a PR

1. Read **[docs/coding-style.md](docs/coding-style.md)** — layers (`cmd` → `usecase` → `repository` → `storage` / `feeds`), `slog`, and ingest conventions. **Automated agents:** follow **[AGENTS.md](AGENTS.md)** so changes stay consistent with this doc.
2. Follow **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** in issues, reviews, and discussions.
3. If you change [docs/schemas/](docs/schemas/), run **`make contracts`** and commit generated packages under `scrape/contract/`, `pipeline/contract/`, `graph/contract/`.
4. Run tests in the **layer** you touched: `make test-scrape`, `make test-pipeline`, `make test-graph` (from repo root), or `cd scrape/scrape_worker && go build .` (same pattern for `pipeline/pipeline_worker`, `graph/ingest_worker`).
5. **Graph read path:** `graph/api` and `graph/mcp` must not import NATS or scrape packages.
6. If you change Compose or env vars, update **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)** and **[deploy/](deploy/)** in the same PR.

## Commits and branches

- Prefer descriptive commit messages (what changed and why).
- Rebase or merge with your team’s usual Git workflow; avoid force-pushing shared branches without agreement.

## Licensing

By contributing, you agree that your contributions are licensed under the **MIT License** ([LICENSE](LICENSE)), the same as the rest of the repository. See [SECURITY.md](SECURITY.md) for responsible disclosure.
