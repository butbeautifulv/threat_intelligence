---
name: Engage tools full coverage (158 → full port)
overview: "Полный порт HexStrike: 158/158 executable (HTTP+MCP). Нет permanent_N/A — heavy stack в engage-runner-full."
todos:
  - id: p9a-metrics-docs
    content: "P9a: Метрики — 158 catalog / live / bridge / subprocess; README KPI full port"
    status: pending
  - id: p9b-unified-dispatch
    content: "P9b: Единый ToolExecutor — bridge/playbook/agent до Runner; HTTP + MCP"
    status: in_progress
  - id: p9c-bridge-coverage
    content: "P9c: 55 bridge_api — handlers/playbooks; test-engage-bridge-coverage"
    status: in_progress
  - id: p9d-runner-breadth
    content: "P9d: runner tier-1 breadth + tools.live (57 runner_N/A → live)"
    status: in_progress
  - id: p9g-heavy-runner
    content: "P9g: Heavy stack в runner — burp, ghidra, hashcat, john, gdb, msf, angr, r2, volatility, wpscan (12 tools)"
    status: in_progress
  - id: p9f-ci-gate
    content: "P9f: test-engage-executable-matrix — 158/158 (subprocess+bridge)"
    status: pending
isProject: false
---

# Engage — полный порт каталога (158/158)

## Политика покрытия (operator)

**Нет исключений permanent_N/A.** Все 158 имён каталога должны быть **исполняемы** (MCP + HTTP):

| Путь | Инструменты |
|------|-------------|
| **bridge** | ~55 intel / CTF / BB / agent (in-process) |
| **subprocess (runner)** | ~103 CLI — включая «тяжёлые» 12 |

### 12 бывших permanent_N/A → P9g (headless в runner)

| Binary | Catalog tool(s) | Установка (target) |
|--------|-----------------|-------------------|
| `burpsuite` | burpsuite_scan, burpsuite_alternative_scan | Community JAR + wrapper CLI |
| `ghidra` | ghidra_analysis | Ghidra + `analyzeHeadless` |
| `hashcat` | hashcat_crack | apt `hashcat` |
| `john` | john_crack | apt `john` |
| `hydra` | hydra_attack | apt `hydra` (уже в runner) |
| `gdb` | gdb_analyze, gdb_peda_debug | apt `gdb` (+ peda optional) |
| `metasploit` | metasploit_run | `msfconsole -q -x` wrapper или msfvenom batch |
| `angr` | angr_symbolic_execution | pip `angr` + thin Python driver script |
| `radare2` | radare2_analyze | apt `radare2` |
| `volatility` | volatility_analyze | pip volatility3 + wrapper |
| `wpscan` | wpscan_analyze | gem или официальный Docker-stage copy binary |

Образ: расширить [runner.Dockerfile](../../deploy/engage/docker/runner.Dockerfile) или профиль **`engage-runner-full`** ([compose.runner.yml](../../deploy/engage/compose.runner.yml)).

**Удалить:** `PERMANENT_NA_BINARIES` из [generate-tools-na-matrix.py](../../scripts/engage/generate-tools-na-matrix.py) — классификация только `live` | `bridge_api` | `runner_pending`.

---

## Почему сейчас «113 live»

| Метрика | Значение |
|---------|----------|
| **158** | `tools.yaml` — все имена |
| **113** | `tools.live.yaml` — subprocess enabled (lab slice) |
| **~55** | bridge — MCP only сегодня; HTTP после P9b |
| **~57+12** | не в live / считались permanent — **закрываем P9d+P9g** |

---

## Целевое покрытие

| Цель | Число | Критерий |
|------|-------|----------|
| **Catalog parity** | 158 | `make test-engage-parity` |
| **Executable (full port)** | **158** | `make test-engage-executable-matrix` — каждый tool success/stub |
| **Subprocess live** | **~103** | все non-bridge с бинарником в runner-full |
| **Bridge** | **55** | `make test-engage-bridge-coverage` |

~~P9e permanent_N/A~~ — **отменено** (заменено P9g).

---

## P9b — Unified dispatch

**Branch:** `engage/p9b-unified-tool-dispatch`

HTTP `POST /api/tools/{name}` = MCP `tools/call` (playbook → agent → bridge → runner).

---

## P9c — Bridge (55)

**Branch:** `engage/p9c-bridge-tool-coverage`

---

## P9d — Runner breadth (tier-1 + tier-2 CLI)

**Branch:** `engage/p9d-runner-burp-hydra` / `engage/p9d-runner-breadth`

exiftool, hakrawler, graphql, … + regen `tools.live.yaml`.

---

## P9g — Heavy runner (12 tools) — **full port**

**Branch:** `engage/p9g-runner-heavy-full`

- [ ] `deploy/engage/docker/runner-heavy.Dockerfile` OR multi-stage в runner.Dockerfile с ARG `RUNNER_PROFILE=full`
- [ ] Wrappers в `deploy/engage/docker/wrappers/` для headless-only
- [ ] `RUNNER_BINARIES` + `generate-tools-live.py` — все 12 binary
- [ ] `LookupBinary` allowlist
- [ ] Docs: image size, RAM, `ENGAGE_RUNNER_IMAGE=…-full`

**DoD:** 12/12 в `list-runner-binaries.sh`; catalog tools enabled in live; matrix **0 permanent_N/A**.

---

## P9f — CI gate 158/158

**Branch:** `engage/p9f-executable-matrix-ci`

После P9b+c+d+g.

---

## Параллельность

```text
P9b ──► P9c
  └──► P9d ∥ P9g (heavy Dockerfile touch-disjoint from dispatch)
         └──► P9f
P9a anytime
```

## Verification

```bash
make test-engage
make test-engage-parity
make test-engage-bridge-coverage
make test-engage-na-matrix          # expect permanent_N/A count = 0
make test-engage-runner-profile     # full image
make test-engage-executable-matrix  # 158/158
```

---

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P9a | `engage/p9a-tool-metrics-docs` | pending |
| P9b | `engage/p9b-unified-tool-dispatch` | in_progress |
| P9c | `engage/p9c-bridge-tool-coverage` | in_progress |
| P9d | `engage/p9d-runner-burp-hydra` | in_progress |
| P9g | `engage/p9g-runner-heavy-full` | in_progress |
| P9f | `engage/p9f-executable-matrix-ci` | pending |
| ~~P9e~~ | — | cancelled (no permanent N/A) |
