# Coding style (Veil)

Conventions for the three runtime layers — **scrape**, **pipeline**, **graph** — and shared contracts. When in doubt, mirror [scrape/sources/ti](../scrape/sources/ti) (scrape) and [graph/sources/ti](../graph/sources/ti) (graph).

## Design principles

| Principle | Rule in this repo |
|-----------|-------------------|
| **CLEAN CODE** | Small functions, clear names, one level of abstraction; `cmd/` is wiring only |
| **DRY** | Shared fetch/ledger only in `scrape/`; normalize only in `pipeline/`; MERGE only in `graph/`; cross-layer types only via [docs/schemas](schemas/) + codegen |
| **KISS** | One long-running binary per layer at runtime; no speculative abstractions |
| **DDD** | **`internal/domain/` is required** in every source module and worker; domain types must not import Neo4j, NATS, or HTTP clients |

## Repository layout

| Zone | Path | Integration |
|------|------|-------------|
| **Knowledge** | [docs/](.) | Schemas, runtime, contracts — no application Go code |
| **Deploy** | [deploy/scrape](../deploy/scrape/), [deploy/pipeline](../deploy/pipeline/), [deploy/graph](../deploy/graph/) | Compose per layer only |
| **Scrape** | [scrape/](../scrape/) | Publish `scrape.>` (`scrapev1`) |
| **Pipeline** | [pipeline/](../pipeline/) | `scrape.>` → normalize → `ingest.>` (`ingestv1`) |
| **Graph** | [graph/](../graph/) | Consume `ingest.>` → Neo4j; API/MCP read |

Layers communicate **only via NATS** and documented JSON schemas. No Go imports across `scrape/`, `pipeline/`, `graph/`.

**Agents:** treat this file as source of truth; entry point [AGENTS.md](../AGENTS.md).

---

## Three runtime contexts

| Context | Code | NATS | Must not |
|---------|------|------|----------|
| **Scrape** | [scrape/](../scrape/) | Publish `scrape.>` | `ingestv1`, Bolt, normalize |
| **Pipeline** | [pipeline/](../pipeline/) | `scrape.>` → `ingest.>` | HTTP feeds, Bolt |
| **Graph** | [graph/](../graph/) | Consume `ingest.>` | `scrapev1`, feeds, Vitess |

Shared fetch policy (scrape only): [scrape/feeds](../scrape/feeds/), [scrape/ledger](../scrape/ledger/) (`VITESS_DSN`, `SCRAPE_MIN_REFETCH_AFTER`, `SCRAPE_FORCE_REFETCH`).

---

## Layering (required packages)

Dependency direction — no cycles:

```
cmd/                    → wiring only (flags, env, construct usecase, Run)
internal/domain/        → entities, value objects, validation (no I/O)
internal/repository/    → ports (interfaces)
internal/usecase/       → orchestration
internal/feeds/         → outbound HTTP/GitHub (scrape sources only)
storage/                → adapters at module root (Neo4j in graph; pub in layer pub/)
```

`internal/*` is private to the Go module. Code another binary must import lives outside `internal/` (e.g. `storage/`, `scrapesource/`).

**PR checklist:** `cmd` has no Cypher; `usecase` has no NATS subject strings; scrape does not import `ingestv1`; graph does not import `scrapev1`; no imports across layers.

---

## Source module template (scrape)

Each [scrape/sources/&lt;name&gt;/](../scrape/sources/):

| Package | Responsibility |
|---------|----------------|
| `internal/domain/` | Domain entities for this feed |
| `internal/feeds/` | HTTP/GitHub fetch |
| `internal/usecase/` | Orchestration |
| `internal/scrapepub/` | Map domain → `scrapev1` kinds |
| `scrapesource/` | `factory.Register` |

## Graph source template

Each [graph/sources/&lt;name&gt;/](../graph/sources/):

| Package | Responsibility |
|---------|----------------|
| `internal/domain/` | Graph-side domain types (or contract DTOs only) |
| `graph/sources/<name>/ingest/` | MERGE handlers from `ingestv1` payload (graph layer) |
| `storage/neo4j/` | Cypher implementations |

---

## Contracts (schema-first)

- **Source of truth:** [docs/schemas/scrapev1-envelope.json](schemas/scrapev1-envelope.json), [docs/schemas/ingestv1-envelope.json](schemas/ingestv1-envelope.json)
- **Generated Go:** `scrape/contract/scrapev1`, `pipeline/contract/ingestv1`, `graph/contract/ingestv1` via `scripts/gen-contracts.sh`
- Do not hand-edit `*.gen.go` files
- Human contract matrix: [ingest-contract.md](ingest-contract.md)

---

## Naming

- Compose services / binaries: **`snake_case`** (`scrape_worker`, `pipeline_worker`, `ingest_worker`)
- NATS durable consumers: **`snake_case`** (`pipeline_worker`, `ingest_worker`)
- Go module path: `github.com/butbeautifulv/threat_intelligence/...` (unchanged)

---

## Logging and lifecycle

- **`log/slog`** with structured attributes
- Long-running binaries: **`errgroup`** + cancel on **SIGINT/SIGTERM**
- Explicit **timeouts** on HTTP clients

---

## Errors

- Wrap with `%w` across package boundaries; log once at `usecase` or `cmd`
- Do not silently ignore fetch errors

---

## Configuration

- Environment variables with sensible defaults
- Document new env vars in [docs/threatintel-runtime.md](threatintel-runtime.md)

---

## Tests

- Table-driven unit tests for parsing, normalization, idempotency keys, envelope validation
- Neo4j integration tests: optional, build tag `integration`

---

## License

**MIT** — [LICENSE](../LICENSE), [CONTRIBUTING.md](../CONTRIBUTING.md)
