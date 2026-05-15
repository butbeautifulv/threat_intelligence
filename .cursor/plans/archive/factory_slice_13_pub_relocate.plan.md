---
name: Factory slice 13 pub relocate
overview: "Срез 13: scrapepub → ingest/scrape/pub, ingestpub → ingest/pipeline/pub; go.work + replace."
todos:
  - id: move-scrapepub
    content: ingest/scrape/pub (module scrapepub)
    status: completed
  - id: move-ingestpub
    content: ingest/pipeline/pub (module ingestpub)
    status: completed
  - id: go-work-replace
    content: go.work + все go.mod replace
    status: completed
  - id: delete-old-pub
    content: Удалить scrapers/scrapepub, scrapers/ingestpub
    status: completed
isProject: false
---

# Scrape factory slice 13: NATS pub relocate

| Было | Стало |
|------|--------|
| `scrapers/scrapepub/` | [ingest/scrape/pub/](../../ingest/scrape/pub/) |
| `scrapers/ingestpub/` | [ingest/pipeline/pub/](../../ingest/pipeline/pub/) |

Модули сохраняют имена `scrapepub` / `ingestpub` для минимального diff импортов.

## Критерии

- [x] `go.work` указывает на `ingest/*/pub`
- [x] Старые каталоги удалены
- [x] `go test ./ingest/scrape/pub/...` зелёный
