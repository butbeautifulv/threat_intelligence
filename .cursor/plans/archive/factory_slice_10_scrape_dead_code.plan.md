---
name: Factory slice 10 scrape dead code
overview: "Срез 10: удалить неиспользуемый код в scrape-доменах (components, mongo yaml, cue_schemas), обновить scrapers/README NATS."
todos:
  - id: drop-components
    content: Удалить scrapers/vuln|lola/internal/components
    status: completed
  - id: drop-mongo-yaml
    content: Убрать mongoConfig из config.yaml
    status: completed
  - id: drop-cue-schemas
    content: Удалить scrapers/cue_schemas
    status: completed
  - id: readme-nats
    content: scrapers/README — scrape.> / ingest.> таблица
    status: completed
isProject: false
---

# Scrape factory slice 10: dead code в scrape-доменах

Мастер: [repo_cleanup_slices](repo_cleanup_slices_8202be7e.plan.md).

## Выполнено

| Артефакт | Действие |
|----------|----------|
| `scrapers/vuln/internal/components/` | удалён |
| `scrapers/lola/internal/components/` | удалён |
| `mongoConfig` в `config.yaml` | удалён |
| `scrapers/cue_schemas/` | удалён |
| [scrapers/README.md](../../scrapers/README.md) NATS | `scrape.>` / `ingest.>` |

## Критерии

- [x] Нет `components/init.go` в vuln/lola
- [x] Нет `cue_schemas/`
- [x] README отражает два NATS hop
