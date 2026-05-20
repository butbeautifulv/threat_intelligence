---
name: Engage HexStrike behavioral port
overview: "После execution wave (P9f–k): behavioral parity с hexstrike_server.py — exploit gen, python env, golden JSON, cloud matrix, visual."
todos:
  - id: p10a-exploit-parity
    content: "P10a: AIExploitGenerator parity — шаблоны SQLi/XSS/RCE/XXE из legacy subset tests"
    status: completed
  - id: p10b-python-env
    content: "P10b: PythonEnvironmentManager — install_python + execute_python_script в runner"
    status: completed
  - id: p10c-golden-ctf-bb
    content: "P10c: Golden JSON CTF/BB vs .external fixtures (P1 audit backlog)"
    status: completed
  - id: p10d-cloud-runner-matrix
    content: "P10d: Cloud tools matrix — prowler/scout/pacu/terrascan/netexec в runner-full CI"
    status: completed
  - id: p10e-recovery-parity
    content: "P10e: Recovery parity — classify-error + parameter-adjustments vs IntelligentErrorHandler"
    status: completed
  - id: p10f-benchmark-regression
    content: "P10f: engage-hexstrike-parity.sh regression baseline (не 24x KPI)"
    status: completed
isProject: false
---

# HexStrike behavioral port (P10)

**Prerequisite:** P9f–k merged on `main` (158 executable, tools.live sync, runner_N/A=0).

**Scope:** `.external/hexstrike-ai-master/` (~23k LOC) vs `engage/` + `pkg/`.

## What P9 closed (execution)

| Item | Status after P9f–k |
|------|-------------------|
| 158 catalog names callable | P9f gate |
| tools.live 1:1 catalog | P9h |
| runner_N/A = 0 | P9i |
| Docker smoke heavy | P9j |
| Honest docs KPI | P9k |

## What P10 closes (behavior)

| Legacy subsystem (Python) | Engage today | P10 |
|---------------------------|--------------|-----|
| `AIExploitGenerator` + exploit classes | `cve/exploit.go` subset | **P10a** expand + tests |
| `PythonEnvironmentManager` | catalog only | **P10b** runner venv/exec |
| `CTFWorkflowManager` / `BugBountyWorkflowManager` | Go workflows | **P10c** golden vs Python JSON |
| Cloud MCP tools (prowler, scout, pacu, …) | in catalog | **P10d** CI subprocess smoke |
| `IntelligentErrorHandler` | `recovery.Handler` | **P10e** parity tests |
| Benchmark README claims | skip | **P10f** regression script only |

**Out of scope P10:** `ModernVisualEngine` ANSI TUI (keep JSON visual), real LLM (`ai_test_payload` stays stub).

---

## P10a — Exploit generator parity

**Branch:** `engage/p10a-exploit-parity`

- Port template logic from `hexstrike_server.py` (`AIExploitGenerator`, `SQLiExploit`, `XSSExploit`, …) into `engage/serve/internal/usecase/cve/` or `pkg/engage/exploit/`
- `make test-engage-cve` + fixture tests vs legacy output shapes

---

## P10b — Python environment

**Branch:** `engage/p10b-python-env`

- `install_python_package` / `execute_python_script` — runner script + allowlist
- Optional venv under `/tmp/engage/pyenv` in runner image

---

## P10c — Golden CTF/BB

**Branch:** `engage/p10c-golden-ctf-bb`

- Extract or hand-craft golden JSON from legacy responses
- `make test-engage-ctf` / `test-engage-bugbounty` golden parity

---

## P10d — Cloud runner matrix

**Branch:** `engage/p10d-cloud-runner-matrix`

- Extend `smoke-engage-runner-full.sh` for prowler, scout_suite, kube-hunter, terrascan, pacu, netexec (`--help` or dry-run)

---

## P10e — Recovery parity

**Branch:** `engage/p10e-recovery-parity`

- Map `ErrorType` / `RecoveryAction` enums to engage recovery
- Tests for `/api/error-handling/*` vs legacy classify paths

---

## P10f — Benchmark regression

**Branch:** `engage/p10f-benchmark-regression`

- Document `make test-engage-benchmark` as regression-only, not KPI
- Optional: record baseline timings JSON for nmap/nuclei sample

---

## Parallelism

```text
P10a ∥ P10b ∥ P10c ∥ P10d ∥ P10e ∥ P10f
  → merge serial if touch engage/serve/go.mod
```

## Verification

```bash
make test-engage
make test-engage-executable-matrix
make test-engage-ctf
make test-engage-bugbounty
make test-engage-cve
make test-engage-runner-full-smoke
```

---

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P10a | `engage/p10a-exploit-parity` | merged main |
| P10b | `engage/p10b-python-env` | merged main |
| P10c | `engage/p10c-golden-ctf-bb` | merged main |
| P10d | `engage/p10d-cloud-runner-matrix` | merged main |
| P10e | `engage/p10e-recovery-parity` | merged main |
| P10f | `engage/p10f-benchmark-regression` | merged main |

**Next:** [engage_hexstrike_post_p10_signoff.plan.md](engage_hexstrike_post_p10_signoff.plan.md) (Golang sign-off + optional P11).
