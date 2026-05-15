# Ingest contract (ingestv1 + JetStream)

Single place for the **producer → NATS → ingest-worker → Neo4j** boundary. Application code lives in [pkg/ingestv1](../pkg/ingestv1/), publishers in [scrapers/ingestpub](../scrapers/ingestpub/), consumer in [scrapers/ingest-worker](../scrapers/ingest-worker/).

## Envelope

- **Schema:** `Envelope` in [pkg/ingestv1/envelope.go](../pkg/ingestv1/envelope.go) (`schema_version`, `source`, `kind`, `idempotency_key`, `payload` JSON).
- **Validation:** `Envelope.Validate()` and `validateEnvelopeSource()` in [scrapers/ingest-worker/cmd/main.go](../scrapers/ingest-worker/cmd/main.go) enforce `source` ↔ `kind` pairing before apply.

## JetStream (implicit ADR)

| Decision | Choice |
|----------|--------|
| Streams | **One** stream named **`INGEST`** |
| Subjects | **`ingest.>`** (wildcard); publishers use concrete subjects under `ingest.*` |
| Consumer | **One** durable pull consumer (default `NATS_DURABLE=ingest-worker`), filter `NATS_SUBSCRIBE_SUBJECT` (default `ingest.>`) |
| Dedup | **`Nats-Msg-Id`** from envelope **`idempotency_key`** (see [scrapers/ingestpub](../scrapers/ingestpub)) |

Rationale for one stream: simpler ops, one place to retain/replay traffic; scale-out is additional durable consumer names / instances with the same subject filter (coordinate with JetStream consumer semantics).

## Default publish subjects (Compose env)

| Compose service | Env var | Default subject |
|-----------------|---------|-----------------|
| `sbom` | `SBOM_NATS_SUBJECT` | `ingest.appsec.sbom` |
| `coderules` | `CODERULES_NATS_SUBJECT` | `ingest.appsec.coderules` |
| `nuclei` | `NUCLEI_NATS_SUBJECT` | `ingest.appsec.nuclei` |
| `ti` | `TI_NATS_SUBJECT` | `ingest.ti.events` |
| `vuln` | `VULN_NATS_SUBJECT` | `ingest.vuln.events` |
| `lola` | `LOLA_NATS_SUBJECT` | `ingest.lola.events` |
| `ds` | `DS_NATS_SUBJECT` | `ingest.ds.events` |

Subjects must remain under `ingest.>` so they match stream **`INGEST`**.

## Unknown / unsupported messages

| Situation | Behaviour |
|-----------|-----------|
| JSON decode / `Validate()` / `validateEnvelopeSource()` error | Log, **NAK** with delay (retry) |
| **AppSec** kinds in `switch` (`sbom`, `coderules`, `nuclei`) — not matched | Log **unknown kind**, return `nil`, message **ACK** |
| **TI / vuln / lola / ds** routed by `source` to `*/workeringest` | Unsupported kind inside that handler returns **error** → NAK |

**Why ACK unknown AppSec kinds:** forwards compatibility — newer producers can emit kinds before the worker gains handlers; ack avoids blocking the whole durable consumer. **TI/vuln/lola/ds** paths fail closed on unknown kind so contract mistakes surface as retries. When extending domains, add `case` branches in [scrapers/ingest-worker/cmd/main.go](../scrapers/ingest-worker/cmd/main.go) or in the domain `workeringest` handler.

## Kind matrix (producer → worker → storage)

Worker entry: [scrapers/ingest-worker/cmd/main.go](../scrapers/ingest-worker/cmd/main.go). Storage packages are the same MERGE logic as the scrapers’ Neo4j writers used by **`ingest-worker`**.

| `source` | `kind` | Worker path | Neo4j / graph write package |
|----------|--------|-------------|----------------------------|
| `sbom` | `sbom_osv_record` | `main.go` → `sbomSt` | [scrapers/sbom/storage/neo4j](../scrapers/sbom/storage/neo4j) |
| `sbom` | `sbom_ghsa_document` | `main.go` → `sbomSt` | same |
| `coderules` | `coderules_cwe_row` | `main.go` → `crSt` | [scrapers/coderules/storage/neo4j](../scrapers/coderules/storage/neo4j) |
| `coderules` | `coderules_semgrep_yaml` | `main.go` → `crSt` | same |
| `coderules` | `coderules_codeql_ql` | `main.go` → `crSt` | same |
| `nuclei` | `nuclei_template_yaml` | `main.go` → `nuSt` | [scrapers/nuclei/storage/neo4j](../scrapers/nuclei/storage/neo4j) |
| `ti` | `ti_ioc`, `ti_kev_vulnerability`, `ti_report`, `ti_campaign`, `ti_cluster`, `ti_actor`, `ti_link_campaign_ioc`, `ti_link_cluster_campaign`, `ti_link_campaign_actor`, `ti_link_report_mentions_ioc`, `ti_jsonl_record` | `apps.ti` → [ti/workeringest](../scrapers/ti/workeringest) | TI writer setup in [ti/workeringest](../scrapers/ti/workeringest) |
| `vuln` | `vuln_upsert`, `vuln_merge_exploit` | `apps.vuln` → [vuln/workeringest](../scrapers/vuln/workeringest) | same pattern |
| `lola` | `lola_artifact`, `lola_lofts`, `lola_attack_technique`, `lola_attack_tactic`, `lola_merge_tactic_technique`, `lola_merge_subtechnique`, `lola_link_artifacts` | `apps.lola` → [lola/workeringest](../scrapers/lola/workeringest) | same |
| `ds` | `ds_upsert_sigma`, `ds_upsert_yara`, `ds_upsert_atomic`, `ds_upsert_caldera` | `apps.ds` → [ds/workeringest](../scrapers/ds/workeringest) | same |

Literal kind strings are the `ingestv1.Kind*` constants in [pkg/ingestv1/envelope.go](../pkg/ingestv1/envelope.go).

## Related docs

- [docs/threatintel-runtime.md](threatintel-runtime.md) — ports, Compose, NATS, graph-bootstrap.
- [scrapers/README.md](../scrapers/README.md) — per-scraper env and feeds.
- [scrapers/ingest-worker/README.md](../scrapers/ingest-worker/README.md) — consumer env and local run.
