---
name: Factory slice 9 gate tombstones
overview: "Срез 9 (Veil cleanup): formal E2E smoke gate → удалить tombstone ingest-worker и deprecated scrapers/*/cmd → актуализировать docs/README. Малый diff."
todos:
  - id: e2e-smoke
    content: scripts/smoke_scrape_e2e.sh зелёный (compose profile scrape ×2)
    status: pending
  - id: delete-tombstones
    content: Удалить scrapers/ingest-worker и scrapers/*/cmd (7 stubs)
    status: completed
  - id: docs-paths
    content: docs/coding-style.md, README mermaid и layout table
    status: completed
isProject: false
---

# Scrape factory slice 9: E2E gate + tombstones

## Контекст

| Срез | Статус |
|------|--------|
| 1–8 v2 | **done** — три worker, без `legacy/` |
| **9 (этот)** | Gate + tombstones перед dead-code sweep (10+) |

Мастер-план: [repo_cleanup_slices](repo_cleanup_slices_8202be7e.plan.md).

---

## Phase 1 — E2E smoke

```bash
./scripts/smoke_scrape_e2e.sh --up
```

Проверки:

1. JetStream lag `SCRAPE` / `INGEST` → 0
2. `crawl_resource` в MySQL
3. Cypher counts (рост vs пустой граф)
4. API `/health` (если api в profile)

**E2E env:** см. [docs/threatintel-runtime.md](../../docs/threatintel-runtime.md) — profile `scrape`, `SCRAPE_SOURCES` все 7.

Исправлять blockers **до** удаления cmd (чтобы не ломать отладку локального `go run` без замены в README).

---

## Phase 2 — Удалить tombstones

| Артефакт | Действие |
|----------|----------|
| [scrapers/ingest-worker/](../../scrapers/ingest-worker/) | удалить каталог |
| [scrapers/{ti,vuln,lola,ds,sbom,coderules,nuclei}/cmd/](../../scrapers/ti/cmd/) | удалить `cmd/main.go` (и пустой `cmd/`) |

Канонический бинарь scrape: [ingest/scrape/scrape_worker/](../../ingest/scrape/scrape_worker/).

---

## Phase 3 — Документация

| Файл | Правка |
|------|--------|
| [docs/coding-style.md](../../docs/coding-style.md) | `ingest/pipeline/pipeline_worker`, `ingest/graph/ingest_worker`; lifecycle example → `ingest/scrape/scrape_worker` |
| [README.md](../../README.md) | Mermaid: один `scrape_worker`, `pipeline_worker`, `ingest_worker`; убрать ingest-worker из layout |
| [scrapers/README.md](../../scrapers/README.md) | ссылка на scrape_worker вместо per-scraper cmd |

---

## Вне scope

- Удаление `components/`, `cue_schemas` — срез 10
- Перенос graph writers — срез 12

---

## Критерии готовности

- [ ] `scripts/smoke_scrape_e2e.sh` зелёный (или задокументирован blocker в PR)
- [ ] Нет `scrapers/ingest-worker/`, нет `scrapers/*/cmd/`
- [ ] `docs/coding-style.md` и README отражают `*_worker` пути
- [ ] [veil_refactor.plan.md](veil_refactor.plan.md): срез 9 в таблице прогресса
