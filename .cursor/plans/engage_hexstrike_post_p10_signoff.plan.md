---
name: Engage HexStrike post-P10 signoff
overview: "P9 execution + P10 behavioral merged. Подтверждение переезда HexStrike на Go (veil-engage). Остаток — опциональный hardening."
todos:
  - id: signoff-golang-engage
    content: "Зафиксировать sign-off: Flask :8888 out, veil-engage Go only (docs + README)"
    status: completed
  - id: p11-runner-docker-parity
    content: "P11a: executable matrix в CI на engage-runner-full (не только stubs на host)"
    status: pending
  - id: p11-llm-stub-policy
    content: "P11b: документировать ai_* stubs vs real LLM (out of scope)"
    status: pending
  - id: p11-python-golden-e2e
    content: "P11c: optional e2e install_python + execute_python в runner-full smoke"
    status: pending
  - id: p11-decommission-external
    content: "P11d: .external/hexstrike-ai-master — archive-only, CI guard no import"
    status: pending
isProject: false
---

# Post-P10: переезд на Golang — подтверждение

**Base:** `main` после merge P9f–k + P10a–f.

## Sign-off: HexStrike → Veil Engage (Go)

| Критерий | Статус | Доказательство |
|----------|--------|----------------|
| Runtime | **Go** | `engage/serve/cmd/{api,mcp,worker}`, `pkg/engage`, `pkg/exec` |
| Legacy Flask `:8888` | **Decommissioned** | [docs/mcp-agents.md](docs/mcp-agents.md), Phase 30 |
| MCP tools | **158 catalog** | `make test-engage-parity` |
| HTTP routes | **Parity** | `make test-engage-route-parity` (156 legacy → mapped) |
| Dispatch | **Unified** | `tooldispatch` — HTTP + MCP одна цепочка (P9b) |
| Bridge handlers | **54/54** | `make test-engage-bridge-coverage` (`execute_python_script` → subprocess P10b) |
| Callable matrix | **158/158** | `make test-engage-executable-matrix` |
| tools.live | **158 rows, 0 orphan** | `make test-engage-na-matrix`, runner_N/A=0 |
| Subprocess in runner | **104 enabled** + stubs/wrappers | P9i/g + P10b python |
| Behavioral (P10) | **Merged** | exploit templates, recovery, golden CTF/BB, cloud smoke, benchmark regression |
| Shared logic | **pkg/** | decision, report, exec, api, mcp, auth |

**Вывод:** продуктовый **переезд на Golang для Engage/HexStrike выполнен**. Python monolith не является runtime; `.external/hexstrike-ai-master/` — reference-only.

## Что НЕ является «100% clone» Python LOC

- `ModernVisualEngine` (ANSI TUI) — JSON visual в Go
- Real LLM (`ai_test_payload`, часть `ai_*`) — deterministic stubs
- 24× benchmark KPI из legacy README — regression-only (`test-engage-benchmark-regression`)
- Каждый subprocess = полный clone CLI behavior в lab — зависит от runner image / `--help` smoke

## P11 (optional hardening)

### P11a — CI runner-full matrix
Branch: `engage/p11a-ci-runner-full-matrix`  
Запуск `check-executable-matrix` внутри `engage-runner-full` без host stubs.

### P11b — LLM policy doc
Branch: `engage/p11b-llm-stub-policy`  
Таблица `ai_*` tools: stub vs future LLM provider.

### P11c — Python e2e smoke
Branch: `engage/p11c-python-runner-smoke`  
`smoke-engage-runner-full.sh`: `install_python_package` + `execute_python_script` dry-run.

### P11d — External archive guard
Branch: `engage/p11d-external-guard`  
CI: fail if `engage/` imports `.external/hexstrike*`.

## Verification (release gate)

```bash
make test-engage
make test-engage-parity
make test-engage-route-parity
make test-engage-bridge-coverage      # 54/54
make test-engage-executable-matrix    # 158/158
make test-engage-na-matrix
make test-engage-ctf test-engage-bugbounty test-engage-cve
make test-engage-benchmark-regression
```

## Status

| Wave | Status |
|------|--------|
| P9 execution | **done** on main |
| P10 behavioral | **done** on main |
| Golang sign-off | **confirmed** |
| P11 optional | pending |
