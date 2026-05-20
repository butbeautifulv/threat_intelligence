---
name: Engage HexStrike full port (158 executable)
overview: "Честный full port: 158/158 callable (HTTP+MCP), tools.live = catalog names only, runner_N/A=0, CI P9f, Docker smoke."
todos:
  - id: p9b-unified-dispatch
    content: "P9b: unified tooldispatch HTTP+MCP"
    status: completed
  - id: p9c-bridge-coverage
    content: "P9c: bridge 55/55"
    status: completed
  - id: p9g-heavy-runner
    content: "P9g: heavy wrappers in runner"
    status: completed
  - id: p9f-executable-matrix
    content: "P9f: check-executable-matrix + make test-engage-executable-matrix (158/158)"
    status: completed
  - id: p9h-catalog-live-sync
    content: "P9h: tools.live = 1:1 catalog names; убрать 90 orphan synthetic; subprocess enabled"
    status: completed
  - id: p9i-runner-remaining-binaries
    content: "P9i: закрыть ~57 runner_N/A — Dockerfile + RUNNER_BINARIES + LookupBinary"
    status: completed
  - id: p9j-runner-full-smoke
    content: "P9j: Docker smoke engage-runner-full — sample + matrix subprocess"
    status: completed
  - id: p9k-docs-honest-kpi
    content: "P9k: README, engage-audit, engage-legacy-parity — три метрики, не «готово» одним словом"
    status: completed
isProject: false
---

# HexStrike full port — execution wave (post-P9b/c/g)

**Prerequisite on main:** P9b dispatch, P9c bridge 55/55, P9g heavy wrappers (`7cbf556`).

## Problem (honest)

| Metric | Today | Target |
|--------|-------|--------|
| Catalog names | 158 | 158 |
| `tools.live` rows | 136 (90 **orphan** names not in catalog) | **158** rows = catalog only |
| Catalog names `enabled` in live | ~46 | **~103** subprocess `enabled: true` |
| `runner_N/A` in matrix | ~57 | **0** |
| `make test-engage-executable-matrix` | missing | **158/158** |

**Sign-off 2026-05-16** = decommission `:8888` + name/route parity, **not** full subprocess clone.

---

## P9f — Executable matrix CI

**Branch:** `engage/p9f-executable-matrix`

- [ ] `scripts/engage/check-executable-matrix.py` — for each catalog name: `tooldispatch` classify + minimal dispatch (unit) or HTTP against test API
- [ ] Success = `success:true` OR structured bridge stub (not `tool disabled`, not `unknown tool`)
- [ ] `make test-engage-executable-matrix`; wire `.github/workflows/engage.yml`
- [ ] Fail report: `docs/engage/engage-executable-gaps.md` (auto)

**DoD:** gate passes 158/158 on main after P9h+i merged.

---

## P9h — Catalog ↔ live sync

**Branch:** `engage/p9h-catalog-live-sync`

- [ ] Replace `generate-tools-live.py` policy OR add `generate-tools-catalog-overlay.py`:
  - **One row per catalog name** (158)
  - **bridge_api** (~55): `enabled: false` OK (dispatch uses `Get`, not `MustGet`)
  - **subprocess**: `enabled: true`
  - **Drop** synthetic-only names not in `tools.yaml` (nmap_quick_scan clones)
- [ ] `make test-engage-na-matrix`: live count = subprocess enabled count; orphan = 0
- [ ] Update `TestLoadCatalog_productionMergeOrder` threshold

**DoD:** `live in catalog` = 103 subprocess enabled + 55 bridge defined; 0 orphan live rows.

---

## P9i — Runner remaining binaries (~57)

**Branch:** `engage/p9i-runner-remaining-binaries`

- [ ] Parse `docs/engage/engage-tools-na-matrix.md` `runner_N/A` → install list
- [ ] `deploy/engage/docker/runner.Dockerfile` + apt/pip/go; wrappers where needed
- [ ] Extend `RUNNER_BINARIES` in `generate-tools-live.py` / `generate-tools-na-matrix.py`
- [ ] `engage/serve/internal/runner` LookupBinary allowlist
- [ ] Regen matrix: `runner_N/A` = 0

**DoD:** `list-runner-binaries.sh` covers every subprocess catalog binary.

**Touch-disjoint with P9h:** only Dockerfile + runner allowlist + na-matrix script (P9h owns tools.live.yaml).

---

## P9j — Runner-full Docker smoke

**Branch:** `engage/p9j-runner-full-smoke`

- [ ] `scripts/test/smoke-engage-runner-full.sh` — build runner image, exec sample tools (nmap, burpsuite wrapper, hashcat --help, …)
- [ ] `make test-engage-runner-full-smoke` (optional CI job, allow skip without Docker)
- [ ] Document `ENGAGE_RUNNER_PROFILE=full` in deploy/engage/README.md

**DoD:** smoke green when Docker available.

---

## P9k — Honest docs

**Branch:** `engage/p9k-docs-honest-kpi`

- [ ] README: three KPIs (catalog / executable / subprocess-in-runner)
- [ ] [engage-audit-report.md](../../docs/engage/engage-audit-report.md): amend execution row; link P9f
- [ ] [engage-legacy-parity.md](../../docs/engage/engage-legacy-parity.md): replace «80 enabled»
- [ ] AGENTS.md: full port = P9f gate required before claiming «HexStrike execution done»

**DoD:** no doc says «113 live = full port».

---

## Merge order

```text
P9f ∥ P9h ∥ P9i ∥ P9j ∥ P9k  (parallel)
  → merge P9k, P9i, P9h (conflict: tools.live — P9h last)
  → merge P9f
  → orchestrator: regen matrix, test-engage-executable-matrix, push main
```

## Verification (orchestrator after all merges)

```bash
make test-engage
make test-engage-parity
make test-engage-bridge-coverage
make test-engage-na-matrix
make test-engage-executable-matrix   # 158/158
make test-engage-runner-full-smoke   # optional Docker
python3 scripts/engage/generate-tools-na-matrix.py --check
```

---

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P9b | — | done `be1469e` |
| P9c | — | done `4ab733f` |
| P9g | — | done `cdf3d51` |
| P9f | `engage/p9f-executable-matrix` | in_progress |
| P9h | `engage/p9h-catalog-live-sync` | in_progress |
| P9i | `engage/p9i-runner-remaining-binaries` | in_progress |
| P9j | `engage/p9j-runner-full-smoke` | in_progress |
| P9k | `engage/p9k-docs-honest-kpi` | in_progress |
