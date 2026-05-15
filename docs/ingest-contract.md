# Ingest contracts (harvest + commit + JetStream)

Veil uses **two NATS streams** between isolated layers:

```text
scrape/ → scrape.> (harvest) → pipeline/ → ingest.> (commit) → graph/ → Neo4j
```

**Go source of truth:** [pkg/harvest/](../pkg/harvest/), [pkg/commit/](../pkg/commit/). **JSON docs:** [schemas/harvest-envelope.json](schemas/harvest-envelope.json), [schemas/commit-envelope.json](schemas/commit-envelope.json) — update manually when pkg types change.

## harvest (scrape → pipeline)

- **Go:** [pkg/harvest/](../pkg/harvest/)
- **Stream:** `SCRAPE`, subjects `scrape.>`
- **Publisher:** [scrape/harvest/](../scrape/harvest/) (`cmd/scrape_worker`)
- **Consumer:** [pipeline/ned/](../pipeline/ned/) (`cmd/pipeline_worker`)
- **Dedup (optional):** `Nats-Msg-Id` = `content_key`

## commit (pipeline → graph)

- **Go:** [pkg/commit/](../pkg/commit/) (pipeline + graph)
- **Stream:** `INGEST`, subjects `ingest.>`
- **Publisher:** pipeline via [pipeline/connector/](../pipeline/connector/)
- **Consumer:** [graph/ingest/](../graph/ingest/) (`cmd/ingest_worker`)
- **Dedup:** `Nats-Msg-Id` = `idempotency_key`

### TI commit payloads (NED → graph)

- Upsert kinds (`ti_ioc`, `ti_campaign`, …): payload is **already normalized by NED** (`pipeline/pkg/ti/normalize`).
- Graph ingest **does not re-normalize**; Neo4j node `id` for IOC/Actor/Report comes from `idempotency_key` (`ti:ioc:…`, `ti:actor:…`, `ti:report:…`) via [pkg/commit/ti_node.go](../pkg/commit/ti_node.go).
- NVD CWE/CPE parsing runs only in pipeline ([pipeline/pkg/nvd/](../pipeline/pkg/nvd/)); harvest publishes raw `scrape_nvd_page` only.

## Vitess crawl ledger (scrape only)

Persistent on host at `var/veil/ledger/mysql/` (bind mount). HTTP bodies: `var/veil/blobs/`. See [graph-pack.md](graph-pack.md#persistent-crawl-state-varveil).

| Variable | Meaning |
|----------|---------|
| `VITESS_DSN` | MySQL-compatible ledger ([scrape/harvest/internal/ledger](../scrape/harvest/internal/ledger/)) |
| `SCRAPE_MIN_REFETCH_AFTER` | Default `24h` |
| `SCRAPE_FORCE_REFETCH` | `1` = ignore ledger |
| `SCRAPE_CACHE_DIR` | Disk cache root (`/data/cache` in compose → `var/veil/blobs` on host) |

If ledger says skip but cache file is missing, [FetchIfDue](../scrape/harvest/internal/feeds/fetch.go) refetches over HTTP instead of failing silently.

## Deploy

Per-layer Compose: [deploy/scrape](../deploy/scrape/compose.yml), [deploy/pipeline](../deploy/pipeline/compose.yml), [deploy/graph](../deploy/graph/compose.yml). Full stack: include all three or use [deploy/graph/compose.full.yml](../deploy/graph/compose.full.yml).
