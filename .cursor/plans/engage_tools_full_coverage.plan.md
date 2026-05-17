---
name: Engage tools full coverage (158 → executable)
overview: "Почему 113 live ≠ 158 catalog; план довести до полного исполняемого покрытия (146 exec + 12 documented permanent N/A)."
todos:
  - id: p9a-metrics-docs
    content: "P9a: Метрики и docs — 158/113/55/57/12; execution_mode в API; README"
    status: pending
  - id: p9b-unified-dispatch
    content: "P9b: Единый ToolExecutor — bridge/playbook/agent до Runner; HTTP + MCP"
    status: in_progress
  - id: p9c-bridge-coverage
    content: "P9c: 55 bridge_api — handlers/playbooks; tools.bridge.yaml; test bridge coverage"
    status: in_progress
  - id: p9d-runner-expand
    content: "P9d: runner image + tools.live — hydra, burpsuite, +40 CLI; снять permanent_N/A для них"
    status: in_progress
  - id: p9e-permanent-na
    content: "P9e: 12 permanent_N/A — structured error + альтернативы в catalog"
    status: pending
  - id: p9f-ci-gate
    content: "P9f: make test-engage-executable-matrix — 146/146 green"
    status: pending
isProject: false
---

# Engage — полное покрытие каталога (158 tools)

## Почему сейчас «только 113»

Это **не баг** и не «45 сломанных тулов». Три разных числа:

| Метрика | Значение | Что значит |
|---------|----------|------------|
| **Catalog** | **158** | Все имена в `tools.yaml` (parity с legacy MCP + bridge) |
| **Parity** | **150** | Имена из `.external/hexstrike_mcp.py` ⊂ catalog |
| **Live (`enabled: true`)** | **113** | Запись в `tools.live.yaml` — **subprocess** в engage-runner |

Остальные **45** из 158 (113 + 45 = 158):

| Класс | ~шт | Статус в matrix | Поведение сейчас |
|-------|-----|-----------------|------------------|
| **bridge_api** | 55 | workflow / intel / CTF | MCP → `IsIntelBridgeTool` + handlers; **HTTP `POST /api/tools/{name}` → только Runner** → `tool disabled` если не в live |
| **runner_N/A** | 57 | CLI есть в catalog, нет в live | Не в runner image или не выбраны в `generate-tools-live.py` |
| **permanent_N/A** | 12 | burp, ghidra, hashcat… | Намеренно вне Docker runner |

**Корневая путаница:** `enabled` в YAML = «можно вызвать subprocess», а не «тул существует в каталоге». Bridge-тулы **есть в MCP list** (158), но **не исполняются по HTTP** без live.

Регенерация matrix: `python3 scripts/engage/generate-tools-na-matrix.py` → [engage-tools-na-matrix.md](../../docs/engage-tools-na-matrix.md).

## Целевое покрытие

| Цель | Число | Критерий |
|------|-------|----------|
| **A — Catalog parity** | 158 | Уже есть (`make test-engage-parity`) |
| **B — Executable** | **146** | 158 − 12 permanent_N/A; любой вызов (MCP **и** HTTP) возвращает success или осмысленный structured result |
| **C — Subprocess live** | **~130+** | Реальный CLI в runner image + `tools.live.yaml` |
| **D — Full subprocess** | 158 | **Недостижимо** без GUI/MSFramework в контейнере |

**Operator requirement:** **hydra** (already in runner apt) and **Burp Suite** must be in engage-runner — use Burp Suite Community JAR + headless/`burpsuite` CLI wrapper or documented `java -jar` path; remove `burpsuite` from permanent_N/A in `generate-tools-na-matrix.py` once packaged.

## Архитектура (target)

```text
POST /api/tools/{name}  ─┐
MCP tools/call          ─┼─► pkg/engage/execdispatch (new) или engage/usecase/tools.Dispatch
                           │
                           ├─► permanent_N/A → 422 + alternatives[]
                           ├─► bridge (intel/ctf/bb/agent/playbook)
                           ├─► browser proxy → discovery
                           └─► subprocess → pkg/exec + runner
```

**Правило:** `Registry.List()` для агентов = 158; поле `execution_mode`: `subprocess` | `bridge` | `permanent_na`.

---

## P9a — Метрики и документация

**Branch:** `engage/p9a-tool-metrics-docs`

- [ ] README / engage-tools.md: таблица 158 vs 113 vs executable
- [ ] `GET /api/tools` — поля `enabled`, `execution_mode`, `reason`
- [ ] Обновить [engage-audit-report.md](../../docs/engage-audit-report.md) KPI

**DoD:** нет противоречия «113 live = всё что работает».

---

## P9b — Единый dispatch (HTTP = MCP)

**Branch:** `engage/p9b-unified-tool-dispatch`

- [ ] Вынести логику из `mcpserver/call.go` (playbook → agent → intel bridge → runner) в `engage/serve/internal/usecase/tools/dispatch.go`
- [ ] `router.go` `POST /api/tools/{name}` → `Dispatch` вместо прямого `Tools.Run`
- [ ] `MustGet` не блокирует bridge: `Get` + mode check

**DoD:** `analyze_target_intelligence` успешен по HTTP без записи в `tools.live.yaml`.

---

## P9c — Bridge coverage (55)

**Branch:** `engage/p9c-bridge-tool-coverage`

- [ ] Аудит: `scripts/engage/audit-bridge-coverage.py` — каждый `bridge_api` → handler | playbook | agent
- [ ] Добавить недостающие handlers (или playbook aliases) для `get_*`, `create_*`, `execute_*`, `http_*`, …
- [ ] Опционально `tools.bridge.yaml` с `enabled: true` + `execution_mode: bridge` (merge после live)

**DoD:** `make test-engage-bridge-coverage` — 55/55 не возвращают «not mapped».

---

## P9d — Runner expansion (~57 runner_N/A)

**Branch:** `engage/p9d-runner-breadth`

- [ ] Приоритет по `pkg/decision` effectiveness ≥ 0.85 и matrix
- [ ] Расширить [runner.Dockerfile](../../deploy/engage/docker/runner.Dockerfile) (tier-2: exiftool, hakrawler, graphql-scanner, …)
- [ ] `python3 scripts/engage/generate-tools-live.py` + bump live count
- [ ] `make test-engage-runner-profile` strict ≥ N

**DoD:** live enabled ≥ 130; `runner_N/A` ≤ 15 (только exotic).

---

## P9e — Permanent N/A (12)

**Branch:** `engage/p9e-permanent-na-policy`

- [ ] Catalog `description` + `alternatives: [nuclei_scan, …]`
- [ ] Dispatch возвращает JSON `{ permanent_na: true, alternatives: [...] }`

**DoD:** агент получает явную замену, не generic «tool disabled».

---

## P9f — CI executable matrix

**Branch:** `engage/p9f-executable-matrix-ci`

- [ ] `scripts/engage/check-executable-matrix.sh` — dry-run или stub call per tool
- [ ] `make test-engage-executable-matrix` в engage.yml
- [ ] Обновить README: **146 executable** (не 113)

**DoD:** 146/146 pass; 12 permanent_N/A documented skip.

---

## Параллельность

| Parallel | Serial after |
|----------|----------------|
| P9a | — |
| P9b | P9c (bridge needs dispatch) |
| P9d ∥ P9c (после P9b) | P9f after c+d+e |

## Verification

```bash
make test-engage
make test-engage-parity      # 158 catalog names
make test-engage-na-matrix   # matrix counts
make test-engage-bridge-coverage   # new (P9c)
make test-engage-executable-matrix # new (P9f)
make test-engage-runner-profile
```

---

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P9a | `engage/p9a-tool-metrics-docs` | pending |
| P9b | `engage/p9b-unified-tool-dispatch` | pending |
| P9c | `engage/p9c-bridge-tool-coverage` | pending |
| P9d | `engage/p9d-runner-breadth` | pending |
| P9e | `engage/p9e-permanent-na-policy` | pending |
| P9f | `engage/p9f-executable-matrix-ci` | pending |
