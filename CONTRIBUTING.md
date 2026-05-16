# Contributing

Thank you for improving this project. Small, focused changes are easier to review than large mixed diffs.

## Before you open a PR

1. Read **[docs/coding-style.md](docs/coding-style.md)** (architecture, layering, [PR checklist](docs/coding-style.md#pr-checklist)). **Automated agents:** follow **[AGENTS.md](AGENTS.md)** and the agent rules under `.cursor/rules/` (workflow, parallel branches, critic, **documentation actualization**).
2. Follow **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** in issues, reviews, and discussions.
3. If you change [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update matching JSON in [docs/schemas/](docs/schemas/) manually in the same PR.
4. Run tests in the **layer** you touched (from repo root):

   | Area | Commands |
   |------|----------|
   | Scrape | `make test-scrape` |
   | Pipeline | `make test-pipeline` |
   | Graph | `make test-graph`, `make test-graph-serve` |
   | Engage | `make test-engage`, `make test-engage-parity` (catalog changes), `make test-engage-hardening` (security guards) |
   | Platform bus / loop | `make test-platform-p0`, optional `make test-platform-closed-loop` |
   | Deploy (Helm/Ansible) | `make deploy-helm-template`, `make deploy-ansible-check` |

   Docker smokes when relevant: `make test-graph-read-smoke`, `make test-engage-veil-stack-ci`, `make test-engage-events-pipeline`, `make test-platform-full-loop` (heavy).

5. **Graph read path:** `graph/serve` (`api`, `mcp`) must not import NATS or scrape packages.
6. If you change Compose, Terraform, Ansible, Helm, or env vars, update **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**, **[deploy/](deploy/)**, and **[docs/deploy-platform-hybrid.md](docs/deploy-platform-hybrid.md)** when production deploy behavior changes.
7. **Graph pack version:** if you change ingest-affecting code (sources, `pkg/harvest`, `pkg/commit`, schemas), run `./scripts/release/bump-graph-version.sh patch` and `make check-graph-version`. Current defaults live in **[versions.env](versions.env)**.
8. **Documentation (required for agents and recommended for humans):** after merge-worthy changes, update:
   - Phase / master plans in `.cursor/plans/`
   - Runtime docs (`docs/threatintel-runtime.md`, `docs/engage-runtime.md`, platform loop docs)
   - **[README.md](README.md)** when user-facing capabilities, tests, or architecture status change
   - **[CONTRIBUTING.md](CONTRIBUTING.md)** when contributor workflow or test matrix changes
   - **[.github/repo-description.txt](.github/repo-description.txt)** when the one-line GitHub summary should change, then `make sync-github-metadata` (or let CI on `main` sync it)

   Full checklist: [.cursor/rules/veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc).

## Commits and branches

- Prefer descriptive commit messages (what changed and why).
- Rebase or merge with your team’s usual Git workflow; avoid force-pushing shared branches without agreement.

## Licensing

By contributing, you agree that your contributions are licensed under the **MIT License** ([LICENSE](LICENSE)), the same as the rest of the repository. See [SECURITY.md](SECURITY.md) for responsible disclosure.
