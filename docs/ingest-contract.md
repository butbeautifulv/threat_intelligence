# Ingest contracts (scrapev1 + ingestv1 + JetStream)

Veil uses **two NATS streams** between isolated layers:

```text
scrape/ → scrape.> (scrapev1) → pipeline/ → ingest.> (ingestv1) → graph/ → Neo4j
```

**Schema source of truth:** [schemas/scrapev1-envelope.json](schemas/scrapev1-envelope.json), [schemas/ingestv1-envelope.json](schemas/ingestv1-envelope.json). Regenerate Go: `./scripts/gen-contracts.sh` or `make contracts`.

## scrapev1 (scrape → pipeline)

- **Go (scrape layer):** [scrape/contract/scrapev1](../scrape/contract/scrapev1/)
- **Stream:** `SCRAPE`, subjects `scrape.>`
- **Publisher:** [scrape/scrape_worker](../scrape/scrape_worker/)
- **Consumer:** [pipeline/pipeline_worker](../pipeline/pipeline_worker/)
- **Dedup (optional):** `Nats-Msg-Id` = `content_key`

## ingestv1 (pipeline → graph)

- **Go (pipeline):** [pipeline/contract/ingestv1](../pipeline/contract/ingestv1/)
- **Go (graph):** [graph/contract/ingestv1](../graph/contract/ingestv1/)
- **Stream:** `INGEST`, subjects `ingest.>`
- **Publisher:** pipeline via [pipeline/pub](../pipeline/pub/)
- **Consumer:** [graph/ingest_worker](../graph/ingest_worker/)
- **Dedup:** `Nats-Msg-Id` = `idempotency_key`

## Vitess crawl ledger (scrape only)

| Variable | Meaning |
|----------|---------|
| `VITESS_DSN` | MySQL-compatible ledger ([scrape/ledger](../scrape/ledger/)) |
| `SCRAPE_MIN_REFETCH_AFTER` | Default `24h` |
| `SCRAPE_FORCE_REFETCH` | `1` = ignore ledger |

## Deploy

Per-layer Compose: [deploy/scrape](../deploy/scrape/compose.yml), [deploy/pipeline](../deploy/pipeline/compose.yml), [deploy/graph](../deploy/graph/compose.yml). Full stack: include all three or use [deploy/graph/compose.full.yml](../deploy/graph/compose.full.yml).
