# Pipeline layer (NED)

Architecture rules: [docs/coding-style.md](../docs/coding-style.md).

Consumes `scrape.>`, applies **NED** (normalize, enrichment, deduplication), publishes `ingest.>`.

| Module | Path | Role |
|--------|------|------|
| **connector** | [connector/](connector/) | NATS JetStream publish + stream ensure |
| **ned** | [ned/](ned/) | Transform worker (`pipeline_worker`) |

- **Wire types:** [pkg/harvest/](../pkg/harvest/), [pkg/commit/](../pkg/commit/)
- **Build:** `make test-pipeline` or:

```bash
cd pipeline/ned && go build -o bin/pipeline_worker ./cmd/pipeline_worker
```

- **Deploy:** [deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml)

## ned layout

```
ned/
  cmd/pipeline_worker/     # thin main
  internal/
    components/            # NATS DI
    consumer/              # scrape pull loop
    transform/             # route by source
    dedup/                 # publish with idempotency keys
    sources/{ti,vuln,lola,ds,appsec}/
```
