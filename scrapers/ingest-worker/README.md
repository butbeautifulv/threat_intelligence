# ingest-worker

Long-running **JetStream pull consumer** that reads AppSec ingest envelopes from NATS and writes to **Neo4j** using the same Cypher `MERGE` paths as the **`sbom`**, **`coderules`**, and **`nuclei`** scrapers in `INGEST_MODE=direct`.

Use this service when scrapers run with **`INGEST_MODE=nats`** (publish-only). Without the worker, messages accumulate in the stream until a consumer drains them.

## Related code

| Path | Role |
|------|------|
| [cmd/main.go](cmd/main.go) | Connect Neo4j (three stores), NATS JetStream, pull loop, ack/nak |
| [../ingestpub/](../ingestpub/) | Stream creation helper (`EnsureAppSecStream`) used by publishers too |
| [../../pkg/ingestv1/](../../pkg/ingestv1/) | Envelope schema, kinds, payloads |
| [../sbom/storage/neo4j/](../sbom/storage/neo4j/) | OSV / GHSA writes |
| [../coderules/storage/neo4j/](../coderules/storage/neo4j/) | CWE / Semgrep / CodeQL writes |
| [../nuclei/storage/neo4j/](../nuclei/storage/neo4j/) | Nuclei template writes |

## Environment

| Variable | Default | Meaning |
|----------|---------|--------|
| `NEO4J_URI` | `neo4j://localhost:7687` | Bolt URI |
| `NEO4J_USER` / `NEO4J_PASS` / `NEO4J_DB` | `neo4j` / `neo4jpassword` / `neo4j` | Auth |
| `NATS_URL` | `nats://localhost:4222` | In Compose scrape profile: `nats://nats:4222` |
| `NATS_INGEST_STREAM` | `INGEST` | JetStream stream name |
| `NATS_DURABLE` | `ingest-worker` | Durable consumer name |
| `NATS_SUBSCRIBE_SUBJECT` | `ingest.appsec.>` | Pull filter (must match stream subjects) |
| `INGEST_BATCH` | `10` | Max messages per `Fetch` |
| `INGEST_MAX_WAIT` | `5s` | Max wait per fetch batch |

On startup the worker ensures stream **`INGEST`** exists (subjects **`ingest.appsec.>`**) if missing, then binds a durable pull subscription.

## Run locally

Requires Neo4j, NATS with JetStream, and optionally messages from a scraper in `nats` mode:

```bash
cd scrapers/ingest-worker
go run ./cmd
```

From repo root with `go.work` synced.

## Docker Compose

Service **`ingest-worker`** is defined in the root [docker-compose.yml](../../docker-compose.yml) under **`profiles: ["scrape"]`**, alongside **`nats`**. See [docs/threatintel-runtime.md](../../docs/threatintel-runtime.md) for the full service matrix and ports.

Typical queue mode:

```bash
export INGEST_MODE=nats
docker compose --profile scrape up --build -d neo4j nats ingest-worker sbom
```

Tune `INGEST_MODE` per service in an override file if only some publishers use NATS.
