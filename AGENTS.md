# Guidance for automated agents (Cursor, CI bots, etc.)

**Behavioral guidelines (Karpathy + Veil):** [.cursor/rules/veil-karpathy-guidelines.mdc](.cursor/rules/veil-karpathy-guidelines.mdc), skill [.cursor/skills/veil-karpathy-guidelines/SKILL.md](.cursor/skills/veil-karpathy-guidelines/SKILL.md). Upstream reference: [.external/andrej-karpathy-skills-main/](.external/andrej-karpathy-skills-main/) (do not edit).

**Metacognition on errors (5 Whys, Gemba Kaizen, 1% improvement):** [.cursor/rules/veil-agent-kaizen-metacognition.mdc](.cursor/rules/veil-agent-kaizen-metacognition.mdc) — mandatory when tests, CI, smokes, or builds fail; document root cause before the next fix.

**Documentation in the agent chain:** [.cursor/rules/veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc) — after each merge, actualize plans, runtime docs, and descriptions the next agent reads; use structured reasoning in phase plans (constraints, few-shot examples from prior phases, explicit `make` DoD).

## Agent chain (summary)

| Step | Rule / doc |
|------|------------|
| Plan | Master + phase plan in `.cursor/plans/` |
| Implement | [veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc) |
| Review | [veil-agent-critic.mdc](.cursor/rules/veil-agent-critic.mdc) |
| Subagents | [veil-agent-subagents.mdc](.cursor/rules/veil-agent-subagents.mdc), [`.cursor/agents/manifest.yaml`](.cursor/agents/manifest.yaml) |
| Merge | Prompt merge to `main` ([veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc) § Merge discipline) |
| Document | [veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc) — includes **README.md**, **CONTRIBUTING.md**, **`.github/repo-description.txt`** |
| Security frameworks | [veil-agent-security-frameworks.mdc](.cursor/rules/veil-agent-security-frameworks.mdc), [docs/external-security-frameworks.md](docs/external-security-frameworks.md) |
| Agent evaluation | [docs/agent-evaluation-gaia.md](docs/agent-evaluation-gaia.md) — [arXiv:2311.12983](https://arxiv.org/abs/2311.12983); `make test-agent-eval-pilot` / `test-agent-eval-paper`; HF optional |
| Platform P6 refactor | [veil_platform_refactor_p6.plan.md](.cursor/plans/veil_platform_refactor_p6.plan.md) — **done** |
| Platform P7 pkg/domain | [veil_platform_p7_tests_then_pkg_domain.plan.md](.cursor/plans/veil_platform_p7_tests_then_pkg_domain.plan.md) — **done**; `make test-platform-p7` |
| Platform v8 layers | [veil_platform_v8_layers_master.plan.md](.cursor/plans/veil_platform_v8_layers_master.plan.md) — **done** (P8a–i: renames, `pkg/report`, `pkg/decision`, `pkg/exec`, `pkg/api`, `pkg/mcp`, engage slim, discovery browser) |
| Finish | This file § End-of-task checklist |

**Completed program tracks (reference for few-shot plans):** Platform v3–v4 P0–P4b; Engage 24–30 / HexStrike ([engage-audit-report.md](docs/engage-audit-report.md)); P5 hybrid deploy; **P6** infra DRY; **P7** pkg domain + CI; **v8** logical layers ([veil_platform_v8_layers_master.plan.md](.cursor/plans/veil_platform_v8_layers_master.plan.md)). **Critical fix on main:** catalog merge order `tools.yaml` → `tools.live.yaml` (`634e067`) — without it live tools appear disabled at runtime.

## Before you change code

1. **Read and follow [docs/coding-style.md](docs/coding-style.md)** — CLEAN CODE, DRY, KISS, DDD; four isolated contexts (`discovery/`, `pipeline/`, `knowledge/`, `engage/`); domain packages per source; shared wire types in `pkg/`. Before merge, check the [PR checklist](docs/coding-style.md#pr-checklist).
2. **Do not add root `go.work`** or cross-layer Go imports between `discovery/`, `pipeline/`, `knowledge/`, `engage/`. Discovery/pipeline/knowledge integrate via NATS; engage calls knowledge only via HTTP veil-api; all layers may import `pkg/*`.
3. Use **[CONTRIBUTING.md](CONTRIBUTING.md)** for tests; when changing [pkg/harvest/](pkg/harvest/) or [pkg/commit/](pkg/commit/), update [docs/schemas/](docs/schemas/) manually in the same PR.
4. Runtime and deploy: **[docs/threatintel-runtime.md](docs/threatintel-runtime.md)**, **[docs/ingest-contract.md](docs/ingest-contract.md)**, **[deploy/README.md](deploy/README.md)**.
5. Versions: **[versions.env](versions.env)** is the single source of truth for `APP_VERSION` and `GRAPH_PACK_VERSION`.

Reference modules: [discovery/harvest/internal/sources/ti/](discovery/harvest/internal/sources/ti/), [knowledge/ingest/internal/sources/ti/](knowledge/ingest/internal/sources/ti/), [pipeline/ned/internal/sources/ti/](pipeline/ned/internal/sources/ti/).

## Planning and commit rhythm (required for multi-phase work)

Keep diffs reviewable: **one git commit per completed phase or slice**, not one giant commit at the end.

1. **Master plan** — before coding, write or update a master plan (status table with **phase / branch / status / owner**, dependencies). For Engage work, keep plans under `.cursor/plans/` (e.g. `engage_hexstrike_master_*.plan.md`, `engage/engage_phase_*.plan.md`).
2. **Phase plan** — for the active phase only, add or open a slice plan derived from the master plan (scope, files, acceptance).
3. **Branch per stream** — implementers work on `engage/phase-<NN>-<slug>` (or `feat/<layer>-phase-<NN>-<slug>`), not directly on `main` when multiple agents run in parallel. See [.cursor/rules/veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc).
4. **Execute one phase** — implement only what that phase plan covers; run tests for touched layers.
5. **Commit on the branch** — `git add` + commit like `feat(engage): Phase N — <short title>`; `git push -u origin HEAD`; open a PR to `main`.
6. **Critic gate** — the **orchestrator / main agent session** acts as critic & compliance ([.cursor/rules/veil-agent-critic.mdc](.cursor/rules/veil-agent-critic.mdc)): plan scope, architecture, tests, graph version; verdict APPROVE / REQUEST_CHANGES before merge.
7. **Merge to `main` promptly** — after critic APPROVE, merge and `git push origin main` so the repo does not drift across parallel branches. See [veil-agent-parallel-branches.mdc](.cursor/rules/veil-agent-parallel-branches.mdc) § Merge discipline.
8. **Update master plan** — on merge, mark phase `done`, note merge commit SHA; clear or archive branch name.
9. **Actualize documentation** — plans, **[README.md](README.md)**, **[CONTRIBUTING.md](CONTRIBUTING.md)**, **[.github/repo-description.txt](.github/repo-description.txt)** (`make sync-github-metadata`), runtime/deploy docs, parity matrices per [veil-agent-documentation.mdc](.cursor/rules/veil-agent-documentation.mdc); list touched doc paths in the commit or PR.

If the user asks to “stage all” or catch up after many phases, still document phase boundaries in the commit message body.

### Parallel agents (summary)

| Role | Branch | Merge to `main` |
|------|--------|-----------------|
| Implementer (Task / subagent / second chat) | `engage/phase-NN-slug`, `platform/p0-*` | Only after critic APPROVE; do not start next phase until prior merge is on `main` |
| Critic & compliance (default for orchestrator chat) | `main` | Merges approved branches, pushes `main`, then starts next phase |

Independent phases may run on **different branches at the same time** only if merges keep pace; otherwise **serialize merges** to avoid divergence. Serial phases rebase onto `main` after dependencies merge.

## End-of-task checklist (required)

Complete every step that applies before you consider the task done:

1. **Tests** — run layer targets from repo root: `make test-discovery`, `make test-pipeline`, `make test-knowledge`, `make test-engage` for the layers you touched. For `knowledge/serve` only: `make test-knowledge-serve` (`-race`). Knowledge read Docker smoke: `make test-graph-read-smoke`. Engage: `make test-engage-parity` when changing catalog. Engage events bus (`engage/.../events`, `pipeline/engage-events`, `knowledge/ingest/.../engage`): also `make test-pipeline`; Docker `make test-engage-events-pipeline`, `make test-engage-veil-stack-ci`. Platform: `make test-platform-p7` (pkg domain + bus, no Docker), `make test-platform-p0` (bus unit tests), `make test-platform-closed-loop` (Docker pilot), optional `make test-platform-full-loop` (discovery + engage, heavy) — [docs/platform-closed-loop-pilot.md](docs/platform-closed-loop-pilot.md), [docs/platform-full-loop-smoke.md](docs/platform-full-loop-smoke.md). Engage hardening (secured infra): `make test-engage-hardening` — [docs/engage-hardening.md](docs/engage-hardening.md).
2. **Graph version** — if you changed ingest-affecting paths (`discovery/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `knowledge/ingest/internal/sources/` including `engage/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`), run `./scripts/release/bump-graph-version.sh patch` and rebuild/publish the graph pack when a new ZIP is needed.
3. **Pre-commit check** — `./scripts/release/check-graph-version-bump.sh` (or `make check-graph-version`).
4. **Commit** — descriptive message (what changed and why). Do not commit secrets or `data/`. Use `git add -A -- . ':!data'` when `data/` causes permission errors. Exclude `**/__pycache__/`.
5. **Push** — `git push origin HEAD` unless the user explicitly forbade push or there is no remote.
6. **GitHub description** — if [.github/repo-description.txt](.github/repo-description.txt) changed, run `make sync-github-metadata` (or rely on [`.github/workflows/docs.yml`](.github/workflows/docs.yml) on push to `main`).

## Graph pack releases

- Default version: see [versions.env](versions.env).
- Workflow: [docs/graph-pack.md](docs/graph-pack.md).
- Publish: `GRAPH_PACK_VERSION=vX.Y.Z ./scripts/release/publish-graph-pack.sh`.
