---
name: Factory slice 12 graph promote
overview: "Срез 12: graph ingest под scrapers/*/graph/ingest + ingest/graph/workeringest; ingest_worker без прямых scrapers imports."
todos:
  - id: graph-ingest-packages
    content: scrapers/*/graph/ingest (apply + setup)
    status: completed
  - id: ingest-workeringest
    content: ingest/graph/workeringest/{ti,vuln,lola,ds}
    status: completed
  - id: ingest-worker-wiring
    content: ingest_worker → workeringest/* only
    status: completed
  - id: drop-workeringest
    content: Удалить scrapers/*/graph/workeringest
    status: completed
isProject: false
---

# Scrape factory slice 12: graph promote

## Архитектура

```text
ingest/graph/ingest_worker
  → ingest/graph/workeringest/{ti,vuln,lola,ds}
    → scrapers/*/graph/ingest (MERGE + internal)
    → scrapers/*/graph/neo4j
```

`ingest_worker` **не** импортирует `scrapers/*/graph/workeringest` напрямую.

## Критерии

- [x] `ingest/graph/workeringest/*` существует
- [x] `scrapers/*/graph/workeringest` удалён
- [x] `go build ./ingest/graph/ingest_worker/...` зелёный

**Дальше (срез 13+):** физический перенос MERGE-кода из `scrapers/*/graph` в `ingest/graph/storage/*` — отдельное решение (Go `internal`).
