---
name: Factory slice 11 scrapev1 only
overview: "Срез 11: scrape-слой публикует только scrapev1 DTO; ingestv1 строится в pipeline_worker."
todos:
  - id: scrapev1-payloads
    content: TIKEVRow, TIJSONLLine, VulnMergeExploit, Lola* в pkg/scrapev1
    status: completed
  - id: scrapepub-domains
    content: ti/vuln/lola internal/scrapepub без pkg/ingestv1
    status: completed
  - id: pipeline-handlers
    content: handle/ti,vuln,lola — map scrapev1 → ingestv1
    status: completed
isProject: false
---

# Scrape factory slice 11: scrapev1-only payloads

## Изменения

- [pkg/scrapev1/envelope.go](../../pkg/scrapev1/envelope.go) — raw DTO для KEV, JSONL, exploit merge, MITRE rows
- `scrapers/{ti,vuln,lola}/internal/scrapepub` — только `pkg/scrapev1`
- [ingest/pipeline/pipeline_worker/internal/handle/](../../ingest/pipeline/pipeline_worker/internal/handle/) — конвертация в `ingestv1`

## Критерии

- [x] Scrape-домены не импортируют `pkg/ingestv1` в scrapepub
- [x] `go test ./ingest/pipeline/pipeline_worker/internal/handle/...` зелёный
