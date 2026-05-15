# ingest-worker

Long-running **JetStream pull consumer** that reads **`ingestv1`** envelopes from NATS and writes to **Neo4j** using the same semantics as the scrapers’ **`storage/neo4j`** packages:

- **AppSec:** `sbom`, `coderules`, `nuclei` (subjects `ingest.appsec.*`)
- **TI:** `ti` feeds / JSONL (`ingest.ti.*` or default `ingest.ti.events`)
- **Graph scrapers:** `vuln`, `lola`, `ds` (default subjects `ingest.vuln.events`, `ingest.lola.events`, `ingest.ds.events`)

The binary follows the same lifecycle pattern as other long-running scrapers: **`golang.org/x/sync/errgroup`** with a context cancelled on **SIGINT/SIGTERM** (see [docs/coding-style.md](../../docs/coding-style.md)). Envelope **`source`** must match **`kind`**; unknown kinds are logged and **acked** so newer producers do not block the consumer (see [docs/ingest-contract.md](../../docs/ingest-contract.md) for the full matrix and ack vs NAK rules).

Use this service with the **`scrape`** profile whenever producers publish to JetStream. Without the worker, messages accumulate in the stream until a consumer drains them.

## Related code

| Path | Role |
|------|------|
| [cmd/main.go](cmd/main.go) | Neo4j writers (sbom, coderules, nuclei, ti, vuln, lola, ds), NATS JetStream, pull loop, ack/nak |
| [../ingestpub/](../ingestpub/) | Stream ensure (`EnsureIngestStream`, `ingest.>`) used by publishers too |
| [../../pkg/ingestv1/](../../pkg/ingestv1/) | Envelope schema, kinds, payloads |
| [../../docs/ingest-contract.md](../../docs/ingest-contract.md) | Subject defaults, kind→handler matrix, JetStream ADR |
| [../ti/workeringest/](../ti/workeringest/) | TI apply path (importable from outside `ti/internal`) |
| [../vuln/workeringest/](../vuln/workeringest/) | Vulnerability apply path |
| [../lola/workeringest/](../lola/workeringest/) | Lola apply path |
| [../ds/workeringest/](../ds/workeringest/) | Detection-content apply path |
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
| `NATS_SUBSCRIBE_SUBJECT` | `ingest.>` | Pull filter (must match stream subjects) |
| `INGEST_BATCH` | `10` | Max messages per `Fetch` |
| `INGEST_MAX_WAIT` | `5s` | Max wait per fetch batch |

On startup the worker ensures stream **`INGEST`** exists with subjects **`ingest.>`** (widens legacy `ingest.appsec.>`-only streams on update).

If you previously used **`NATS_SUBSCRIBE_SUBJECT=ingest.appsec.>`**, switch to **`ingest.>`** (or a narrower filter) so TI messages are delivered; you may need a new **`NATS_DURABLE`** name if the old consumer filter cannot be updated in place.

## Run locally

Requires Neo4j, NATS with JetStream, and messages from a scraper:

```bash
cd scrapers/ingest-worker
go run ./cmd
```

From repo root with `go.work` synced.

## Compose (scrape profile)

```bash
docker compose --profile scrape up --build -d neo4j nats ingest-worker sbom
```

Add **`ti`**, **`vuln`**, **`lola`**, or **`ds`** alongside **`ingest-worker`** for queue-backed ingest of those domains.
