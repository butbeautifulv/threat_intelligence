# Deploy and scale

Per-layer compose lives under [deploy/](../deploy/). See [deploy/README.md](../deploy/README.md) for the full stack and worker scaling.

## Full stack

```bash
./scripts/compose-up-full.sh
```

## Scale pipeline and ingest workers

JetStream pull consumers share a **durable name**; multiple replicas compete for messages (safe scale-out).

```bash
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/compose-up-full.sh
```

Or manually:

```bash
docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml \
  up -d --scale pipeline_worker=3 --scale ingest_worker=3
```

Do **not** scale `scrape_worker` with the same `SCRAPE_SOURCES` on multiple replicas. Use `SCRAPE_WORKER_PARTITION=1` (see [deploy/compose.scale.yml](../deploy/compose.scale.yml)) or split `SCRAPE_SOURCES` per container.

## Graph-only (API + Neo4j)

```bash
docker compose -f deploy/graph/compose.yml up --build -d
```

Root [docker-compose.yml](../docker-compose.yml) is a thin include of the graph layer.

## E2E smoke

```bash
./scripts/smoke_scrape_e2e.sh --up
PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/smoke_scrape_e2e.sh --up
```

## Graph pack release

See [threatintel-runtime.md](threatintel-runtime.md) and [scripts/build-graph-pack.sh](../scripts/build-graph-pack.sh).
