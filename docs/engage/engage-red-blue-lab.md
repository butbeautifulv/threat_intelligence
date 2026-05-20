# Engage red-vs-blue lab (same host, two instances)

This document defines a **controlled lab** topology: a **victim** `engage-api` instance (target) and an **attacker** side that runs an **automated abuse harness** against the victim’s HTTP API. Install + pentest field results: [engage-lab-pentest.md](engage-lab-pentest.md). Historical plan: [`.cursor/plans/archive/engage_userfriendly_install_73f6d9c0.plan.md`](../.cursor/plans/archive/engage_userfriendly_install_73f6d9c0.plan.md).

## Authorization and scope (non-negotiable)

- Run only against **systems you own** or are **explicitly authorized** to test (lab VM, VPN segment, localhost).
- Bind the victim to **loopback** or a **private lab** address. Do not expose raw lab APIs to the public internet without edge controls and a written threat model.
- The harness performs **high-volume, malformed, and concurrent** HTTP traffic against **your victim instance only**. It does **not** justify probing third parties, production tenants, or shared infrastructure without a signed engagement letter.
- “Aggressive” here means **reliability and abuse testing** (fuzz bodies, concurrency, rate stress), not carte blanche for illegal activity.

## Topology

| Role | Default listen | Purpose |
|------|----------------|---------|
| **Victim (B)** | `:8891` | Catalog + tool runner; receives all harness traffic via `ENGAGE_VICTIM_URL`. |
| **Attacker (A)** | `:8890` | Optional second `engage-api` for future MCP/agent-driven flows; isolated workdirs so jobs/audit/files do not collide with the victim. |
| **Harness** | (no listen) | Shell + `curl` (and `jq` if present) driving HTTP against **only** the victim base URL. |

Default ports match the lab plan; override with `ENGAGE_API_LISTEN` before starting an instance if `8890`/`8891` are taken.

State directories default under `/tmp/engage-lab/<role>/` (override with `ENGAGE_LAB_ROOT`). Set at least:

- `ENGAGE_RUNNER_WORKDIR`
- `ENGAGE_FILES_DIR`
- `ENGAGE_AUDIT_DIR`
- `ENGAGE_JOBS_DIR`

so two processes never share `/tmp/engage` or `/var/veil/engage/*` on the same host.

## Harness behavior (what “super aggressive” means)

All of the following are aimed at **`ENGAGE_VICTIM_URL`** (victim only):

- Concurrent `GET /health` and `GET /api/tools` bursts.
- `POST /api/tools/{name}` with **invalid JSON**, **wrong `Content-Type`**, and **oversized** bodies.
- Parallel workers (`xargs -P` or background jobs) to stress connection handling.
- If the victim runs with **auth enabled**, fuzz **only** with **lab-issued** credentials (for example a throwaway token in env), never production secrets.

## Multi-agent roles

| Role | Typical agent id | Branch pattern | Responsibility |
|------|------------------|----------------|----------------|
| Orchestrator / Critic | (main chat) | reviews on `main` | Merge order, risk sign-off, Go/No-Go. |
| Install slice | `engage-implementer` | `engage/install-pNN-<slug>` | Package map, installer, preflight. |
| Lab slice | `engage-implementer` | `engage/lab-pNN-<slug>` | Instance launchers + red harness. |
| Docs slice | `docs-only` or implementer | `docs/engage-install-pNN-<slug>` | Runbooks and README links. |
| Operator | You | — | Field run, logs, bug entries in [`engage-red-blue-bugs.md`](engage-red-blue-bugs.md). |

Related automation rules: [`.cursor/rules/veil-agent-parallel-branches.mdc`](../.cursor/rules/veil-agent-parallel-branches.mdc), [`.cursor/rules/veil-agent-critic.mdc`](../.cursor/rules/veil-agent-critic.mdc). Subagent bindings: [`.cursor/agents/manifest.yaml`](../.cursor/agents/manifest.yaml) (phases `engage-install-lab-p01` … `p08`).

## Merge order (lab + install track)

1. Legal + this runbook (+ README links): **no executable lab code**.
2. `scripts/ops/engage-tools-packages.yaml` (data only).
3. `scripts/ops/install-engage-host-tools.sh` (PM detect, `--plan` / `--yes`).
4. `scripts/engage/preflight-client-tools.sh` (profiles, `--json`).
5. `scripts/engage/run-client-native-api-instance.sh` **victim** wiring (ports + dirs).
6. Same script **attacker** role (or equivalent env contract).
7. `scripts/test/smoke-engage-red-vs-blue.sh` (harness).
8. `Makefile` targets + [`engage-install-linux.md`](engage-install-linux.md) + scripts index.

After PR8: operator field test → micro fix PRs → [`engage-client-native-viability.md`](engage-client-native-viability.md) Go/No-Go.

## Quick start (after scripts land)

```bash
# Terminal 1 — victim
./scripts/engage/run-client-native-api-instance.sh victim

# Terminal 2 — optional attacker API (same repo, isolated dirs)
./scripts/engage/run-client-native-api-instance.sh attacker

# Terminal 3 — harness (victim must be up)
export ENGAGE_VICTIM_URL=http://127.0.0.1:8891
chmod +x ./scripts/test/smoke-engage-red-vs-blue.sh
./scripts/test/smoke-engage-red-vs-blue.sh
```

Install tooling on the host first: [`engage-install-linux.md`](engage-install-linux.md) and [`engage-client-dependencies.md`](engage-client-dependencies.md).
