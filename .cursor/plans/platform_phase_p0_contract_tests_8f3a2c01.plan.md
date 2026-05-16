---
name: Platform P0 Contract Tests
overview: "P0: ingestPublisher interface, consumer unit tests, validateEnvelopeSource, NATS ensure integration, make test-platform-p0."
todos:
  - id: p0-dedup-publisher-iface
    content: "IngestPublisher interface in dedup for test doubles"
    status: completed
  - id: p0-pipeline-consumer
    content: "pipeline/ned/internal/consumer consumer_test.go"
    status: completed
  - id: p0-graph-ingest
    content: "graph/ingest validateEnvelopeSource + handleMsg engage"
    status: completed
  - id: p0-nats-ensure
    content: "pipeline/connector EnsureIngestStream integration test"
    status: completed
  - id: p0-make
    content: "make test-platform-p0 in Makefile"
    status: completed
isProject: false
---

# Platform P0 — contract & consumer tests

**Ветка:** `platform/p0-contract-consumer-tests`

## Deliverables

1. `dedup.IngestPublisher` — `ProcessScrapeMessage` testable without live NATS
2. `consumer_test.go` — invalid JSON / validate errors; TI harvest → ingest publish recorded
3. `ingest/consumer_test.go` — `validateEnvelopeSource` table; Engage tool run → Apply.Engage
4. `connector/nats/publish_test.go` — embedded NATS, `EnsureIngestStream`
5. `make test-platform-p0`

## Verify

```bash
make test-platform-p0
make test-pipeline
make test-graph
```
