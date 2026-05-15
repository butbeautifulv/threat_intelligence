# pipeline_worker

Context 2: pull **`scrape.>`**, normalize and deduplicate, publish **`ingestv1`** to **`ingest.>`**.

```bash
cd pipeline/pipeline_worker
export NATS_URL=nats://localhost:4222
export NATS_SCRAPE_SUBSCRIBE_SUBJECT='scrape.>'
export DS_INGEST_SUBJECT=ingest.ds.events
# ... TI_INGEST_SUBJECT, VULN_INGEST_SUBJECT, etc.
go run ./cmd
```

Compose: service **`pipeline_worker`** in [deploy/pipeline/compose.yml](../../deploy/pipeline/compose.yml). Full stack: `./scripts/compose-up-full.sh`.
