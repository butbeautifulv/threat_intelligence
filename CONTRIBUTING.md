# Contributing

Thank you for improving this project. Small, focused changes are easier to review than large mixed diffs.

## Before you open a PR

1. Read **[docs/coding-style.md](docs/coding-style.md)** (architecture, layering, [PR checklist](docs/coding-style.md#pr-checklist)). **Automated agents:** follow **[AGENTS.md](AGENTS.md)** and `.cursor/rules/`.
2. Follow **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)**.
3. If you change [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update [docs/schemas/](docs/schemas/) in the same PR.
4. **Tests** — run targets for layers you touched. Full matrix: **[README.md#tests](README.md#tests)**.

   **PR minimum (typical):**

   | Area | Command |
   |------|---------|
   | Layer you changed | `make test-discovery` / `test-pipeline` / `test-knowledge` / `test-engage` |
   | Catalog change | `make test-engage-parity` |
   | Security guards | `make test-engage-hardening` |
   | Platform pkg/bus | `make test-platform-p7` |
   | Graph read path only | `make test-knowledge-serve` |

   Docker smokes when relevant: `make test-graph-read-smoke`, `make test-engage-events-pipeline`, `make test-platform-unified-edge`.

5. **Graph read path:** `knowledge/serve` must not import NATS or scrape packages.
6. Compose / deploy changes: update **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)** and **[deploy/](deploy/)** as needed.
7. **Graph pack version:** ingest-affecting changes → `./scripts/release/bump-graph-version.sh patch` and `make check-graph-version` ([versions.env](versions.env)).
8. **Documentation:** update runtime docs and **[README.md](README.md)** when user-facing behavior changes; completed plans go to **[.cursor/plans/archive/](.cursor/plans/archive/)**; refresh **[.github/repo-description.txt](.github/repo-description.txt)** + `make sync-github-metadata` when the one-line summary changes.

   Checklist: [.cursor/rules/veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc).

## Commits and branches

- Prefer descriptive commit messages (what changed and why).
- Rebase or merge with your team’s usual Git workflow; avoid force-pushing shared branches without agreement.

## Licensing

Contributions are licensed under the **MIT License** ([LICENSE](LICENSE)). See [SECURITY.md](SECURITY.md) for responsible disclosure.
