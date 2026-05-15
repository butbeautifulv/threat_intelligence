# Pipeline layer

Consumes `scrape.>`, normalizes to `ingestv1`, publishes `ingest.>`.

- **Worker:** [pipeline_worker/](pipeline_worker/)
- **Normalize:** [internal/normalize/](internal/normalize/)
- **Contract:** [contract/ingestv1/](contract/ingestv1/)
- **Build:** `cd pipeline && go build ./pipeline_worker/...`
- **Deploy:** [../deploy/pipeline/compose.yml](../deploy/pipeline/compose.yml)
