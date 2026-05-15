# Contributing

Thank you for improving this project. Small, focused changes are easier to review than large mixed diffs.

## Before you open a PR

1. Read **[docs/coding-style.md](docs/coding-style.md)** (architecture, layering, [PR checklist](docs/coding-style.md#pr-checklist)). **Automated agents:** follow **[AGENTS.md](AGENTS.md)**.
2. Follow **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** in issues, reviews, and discussions.
3. If you change [pkg/harvest](pkg/harvest/) or [pkg/commit](pkg/commit/), update matching JSON in [docs/schemas/](docs/schemas/) manually in the same PR.
4. Run tests in the **layer** you touched: `make test-scrape`, `make test-pipeline`, `make test-graph` (from repo root), or `cd scrape/harvest && go build ./cmd/scrape_worker` (same pattern for `pipeline/ned/cmd/pipeline_worker`, `graph/ingest/cmd/ingest_worker`).
5. **Graph read path:** `graph/serve` (`api`, `mcp`) must not import NATS or scrape packages.
6. If you change Compose or env vars, update **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)** and **[deploy/](deploy/)** in the same PR.

## Commits and branches

- Prefer descriptive commit messages (what changed and why).
- Rebase or merge with your team’s usual Git workflow; avoid force-pushing shared branches without agreement.

## Licensing

By contributing, you agree that your contributions are licensed under the **MIT License** ([LICENSE](LICENSE)), the same as the rest of the repository. See [SECURITY.md](SECURITY.md) for responsible disclosure.
