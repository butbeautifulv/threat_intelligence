---
name: Engage Phase 25 Execution 100
overview: "Phase 25 (R125–R130): N/A matrix на 100% catalog, расширение runner, ≥100 live tools, strict matrix в compose CI, triangle artifact."
todos:
  - id: p25-r125-na-matrix
    content: "R125: generate-tools-na-matrix.py + docs/engage-tools-na-matrix.md (158/158)"
    status: completed
  - id: p25-r126-runner
    content: "R126: runner.Dockerfile — jaeles, x8, whatweb, nbtscan, binwalk, enum4linux-ng"
    status: completed
  - id: p25-r127-live-100
    content: "R127: tools.live.yaml ≥100 enabled via generate-tools-live.py"
    status: completed
  - id: p25-r128-strict-ci
    content: "R128: compose smoke strict matrix без || true; min 30"
    status: completed
  - id: p25-r129-triangle
    content: "R129: engage.yml artifact engage-mcp-runner-triangle.csv"
    status: completed
  - id: p25-r130-docs
    content: "R130: engage-tools.md permanent N/A + make test-engage-na-matrix"
    status: completed
isProject: false
---

# Phase 25 — Execution breadth II (R125–R130)

Родитель: [engage_master_post-audit_ec180f8b.plan.md](.cursor/plans/engage_master_post-audit_ec180f8b.plan.md)

**Ветка:** `engage/phase-25-execution-100`

**Зависимость:** Phase 24 желательна (CI e2e); не блокирует catalog/runner work.

**Baseline:** 80 `enabled: true` в [tools.live.yaml](engage/serve/catalog/tools.live.yaml); 158 имён в [tools.yaml](engage/serve/catalog/tools.yaml).

---

## R125 — N/A execution matrix

**Файлы:** [scripts/engage/generate-tools-na-matrix.py](scripts/engage/generate-tools-na-matrix.py), [docs/engage-tools-na-matrix.md](docs/engage-tools-na-matrix.md)

| Status | Meaning |
|--------|---------|
| `live` | `tools.live.yaml` enabled + binary in runner (or `api` bridge) |
| `runner_N/A` | Catalog binary not in runner image; could enable later |
| `bridge_api` | In-process / workflow (`binary: api`, `bugbounty`, …) |
| `permanent_N/A` | GUI / heavy / legacy-only (ghidra, burp, metasploit GUI, angr, wpscan gem stack) |

`make test-engage-na-matrix` — regenerate + assert 158 rows and ≥100 live.

---

## R126 — Runner expansion

**Файл:** [deploy/engage/docker/runner.Dockerfile](deploy/engage/docker/runner.Dockerfile)

| Binary | Install |
|--------|---------|
| jaeles, x8 | `go install` (pd stage) |
| whatweb, nbtscan, binwalk | apt |
| enum4linux-ng | pip |

**Permanent N/A (не в образ):** wpscan (Ruby gem stack), ghidra, burpsuite, metasploit, angr.

---

## R127 — 100+ live tools

**Файл:** [scripts/engage/generate-tools-live.py](scripts/engage/generate-tools-live.py)

- Расширить `RUNNER_BINARIES`, `PREFERRED`, `SYNTHETIC` (Phase 25 block).
- `python3 scripts/engage/generate-tools-live.py` → **≥100** entries.

---

## R128 — Strict matrix CI

**Файлы:** [scripts/test/smoke-engage-compose.sh](scripts/test/smoke-engage-compose.sh), [scripts/engage/list-runner-binaries.sh](scripts/engage/list-runner-binaries.sh)

- `ENGAGE_TOOL_MATRIX_STRICT=1 ENGAGE_TOOL_MATRIX_MIN=30` — **fail** compose smoke on failure (убрать `|| true`).
- Matrix resolves binaries against runner container when `ENGAGE_RUNNER_IMAGE` set.

---

## R129 — Triangle CSV artifact

**Файл:** [.github/workflows/engage.yml](.github/workflows/engage.yml)

- Run `audit-mcp-runner-triangle.py` after catalog steps.
- Upload `docs/engage-mcp-runner-triangle.csv` as artifact.

---

## R130 — Documentation

**Файл:** [docs/engage-tools.md](docs/engage-tools.md)

- Секция **Permanent N/A** (heavy/GUI).
- Ссылка на [engage-tools-na-matrix.md](docs/engage-tools-na-matrix.md).

---

## Definition of Done

- [x] `make test-engage-na-matrix` green (158 rows, 113 live enabled)
- [x] `grep -c 'enabled: true' tools.live.yaml` ≥ 100 (113)
- [x] `make test-engage` + `make test-engage-parity` + `make test-engage-catalog-args` green
- [x] Compose smoke: strict matrix via `ENGAGE_RUNNER_CONTAINER` (Docker CI)
- [x] CI: triangle artifact on `test` job
- [x] Master plan Phase 25 row updated

---

## Verify (Karpathy)

```bash
make test-engage
make test-engage-parity
make test-engage-catalog-args
make test-engage-na-matrix
python3 scripts/engage/generate-tools-live.py
grep -c 'enabled: true' engage/serve/catalog/tools.live.yaml
make test-engage-compose   # Docker, strict matrix
```
