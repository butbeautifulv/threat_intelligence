---
name: Veil Platform v3 Test Then Dedup
overview: "Платформа: тесты на bus/контракты (P0), DRY NATS (P1), closed loop (P2+). Цель — авто-обогащение и улучшение безопасности целевых систем."
todos:
  - id: p0-contract-consumer
    content: "P0: contract + consumer tests (pkg, pipeline, graph ingest)"
    status: completed
  - id: p1-nats-dry
    content: "P1: pkg/natsjet consumer loop, dedupe EnsureStream"
    status: completed
  - id: p2-graph-target-state
    content: "P2: graph read API для engage decisions"
    status: pending
  - id: p3-closed-loop
    content: "P3: policy pilot scrape→graph→engage→events"
    status: pending
isProject: false
---

# Veil Platform v3 — test harness, then dedup

## Vision

| Layer | Role |
|-------|------|
| Scrape | Discovery (internet/feeds) |
| Pipeline | Enrichment (NED) |
| Graph | Knowledge (Neo4j + veil-api) |
| Engage | Action (tools, workflows) |

**Ultimate:** closed loop — discover → enrich → remember → act → learn (events → ingest).

## Architecture constraint

No cross-import `scrape`/`pipeline`/`graph`/`engage`. Shared code only in `pkg/*`.

**Not duplicated:** engage `exec` runner vs scrape HTTP harvest (different mechanisms).

**Duplicated (P1 target):** JetStream ensure/publish/pull loops.

## Phases

| Phase | Branch | DoD |
|-------|--------|-----|
| P0 | `platform/p0-contract-consumer-tests` | merged `8bdf4c4` — `make test-platform-p0` green |
| P1 | `platform/p1-nats-dry` | `pkg/natsjet` Ensure* streams + `RunPullLoop` |
| P2 | `platform/p2-graph-target` | engage uses graph context consistently |
| P3 | `platform/p3-closed-loop` | one target-class pilot documented |
