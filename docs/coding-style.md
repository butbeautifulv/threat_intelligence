# Coding style (Veil)

Conventions for the three runtime layers — **scrape**, **pipeline**, **graph** — and shared contracts. When in doubt, mirror [scrape/harvest/internal/sources/ti/](../scrape/harvest/internal/sources/ti/) (scrape) and [graph/ingest/internal/sources/ti/](../graph/ingest/internal/sources/ti/) (graph write path).

**Agents:** treat this file as source of truth; entry point [AGENTS.md](../AGENTS.md).

## Design principles

| Principle | Rule in this repo |
|-----------|-------------------|
| **CLEAN CODE** | Small functions, clear names, one level of abstraction; `cmd/` is wiring only |
| **DRY** | Shared fetch/ledger only in `scrape/`; normalize only in `pipeline/`; MERGE only in `graph/`; cross-layer types in `pkg/*` (wire envelopes, NATS helpers, TI helpers) |
| **KISS** | One long-running binary per layer at runtime; no speculative abstractions |
| **DDD** | Domain package per source module, no I/O (no Neo4j, NATS, HTTP clients in domain types) |

## Repository map

| Zone | Path | Role |
|------|------|------|
| **Knowledge** | [docs/](./) | Schemas, runtime, contracts — no application Go code |
| **Deploy** | [deploy/](../deploy/) | Compose per layer |
| **Scrape** | [scrape/](../scrape/) | Fetch + ledger → `scrape.>` |
| **Pipeline** | [pipeline/](../pipeline/) | `scrape.>` → NED → `ingest.>` |
| **Graph** | [graph/](../graph/) | Consume `ingest.>` → Neo4j; API/MCP read |
| **Engage** | [engage/](../engage/) | Tool execution, workflows, reports (HTTP API/MCP) |
| **Wire types** | [pkg/](../pkg/) | `harvest`, `commit`, `natsjet`, `ti/*`, `auth`, `engage/*` |

Layers **scrape / pipeline / graph** communicate **only via NATS** and documented JSON schemas. **Engage** calls graph only via **HTTP veil-api** (no Bolt, no NATS). No Go imports across `scrape/`, `pipeline/`, `graph/`, `engage/`. All layers may import `pkg/*`. NVD parse/map lives in [pipeline/pkg/nvd](../pipeline/pkg/nvd/) (pipeline only).

Layer-specific layout, env vars, and build commands:

| Layer | Docs |
|-------|------|
| Scrape | [scrape/README.md](../scrape/README.md), [scrape/harvest/README.md](../scrape/harvest/README.md) |
| Pipeline | [pipeline/README.md](../pipeline/README.md), [pipeline/ned/README.md](../pipeline/ned/README.md) |
| Graph | [graph/README.md](../graph/README.md), [graph/ingest/README.md](../graph/ingest/README.md) |

## Three runtime contexts

| Context | Code | NATS | Must not |
|---------|------|------|----------|
| **Scrape** | [scrape/](../scrape/) | Publish `scrape.>` | `commit`, Bolt, normalize |
| **Pipeline (NED)** | [pipeline/](../pipeline/) | `scrape.>` → `ingest.>` | HTTP feeds, Bolt, MERGE |
| **Graph** | [graph/](../graph/) | Consume `ingest.>` | `harvest`, feeds, Vitess |
| **Engage** | [engage/](../engage/) | — (HTTP to veil-api only) | Bolt, NATS, scrape |

Shared fetch policy (scrape only): [scrape/harvest/internal/feeds](../scrape/harvest/internal/feeds/), [scrape/harvest/internal/ledger](../scrape/harvest/internal/ledger/) (`VITESS_DSN`, `SCRAPE_MIN_REFETCH_AFTER`, `SCRAPE_FORCE_REFETCH`).

Wire contracts: [ingest-contract.md](ingest-contract.md). Go SOT: [pkg/harvest](../pkg/harvest/), [pkg/commit](../pkg/commit/).

---

## Layering (required packages)

Dependency direction — no cycles:

```
cmd/                    → wiring only (flags, env, construct usecase, Run)
domain/                 → entities, value objects, validation (no I/O)
internal/repository/    → ports (interfaces) — where used
internal/usecase/       → orchestration
internal/feeds/         → outbound HTTP/GitHub (scrape sources only)
storage/                → adapters at module root (Neo4j in graph; pub in layer connector)
```

`internal/*` is private to the Go module. Code another binary must import lives outside `internal/` (e.g. `storage/`, `scrapesource/`).

### Domain package paths (by layer)

| Layer | Typical path | Reference |
|-------|--------------|-----------|
| Scrape source | `internal/sources/<name>/internal/domain/` | [scrape/.../ti/internal/domain](../scrape/harvest/internal/sources/ti/internal/domain/) |
| Graph ingest source | `internal/sources/<name>/domain/` | [graph/ingest/.../ti/domain](../graph/ingest/internal/sources/ti/domain/) |
| Graph serve | `internal/domain/` | [graph/serve/internal/domain/](../graph/serve/internal/domain/) |
| Pipeline | `internal/sources/<name>/domain/` when entities exist | [pipeline/ned/.../vuln/domain](../pipeline/ned/internal/sources/vuln/domain/) |

Pipeline/scripts boundary: [scripts/README.md](../scripts/README.md) (`ops/`, `graph-pack/`, `test/`, `housekeeping/` — Neo4j housekeeping is not NED wire dedup). Graph packs: [docs/graph-pack.md](graph-pack.md).

---

## PR checklist

Before merge, verify all items that apply to your layer:

| Rule | Scrape | Pipeline (NED) | Graph |
|------|--------|----------------|-------|
| `cmd/` has no business logic (no HTTP, Cypher, per-source transform) | ✓ | ✓ | ✓ |
| `usecase` has no NATS subject strings | ✓ | ✓ | ✓ |
| No `commit` / Bolt in scrape | ✓ | — | — |
| No `harvest` / feeds / Vitess in graph | — | — | ✓ |
| No cross-layer Go imports | ✓ | ✓ | ✓ |
| NVD parse only in pipeline (`pipeline/pkg/nvd`) | harvest publishes raw page only | ✓ enrich in `sources/vuln/enrich` | ingest does not re-parse NVD |
| Graph ingest does not import `pipeline/pkg/ti/normalize` | — | NED normalizes TI | ✓ |
| Idempotency keys via `pkg/commit` helpers only | — | ✓ | ✓ |
| `graph/serve` does not import NATS or scrape | — | — | ✓ |

### Agent / CI closure

For automated agents and maintainers, before marking work done:

| Step | Command / doc |
|------|----------------|
| Tests (touched layers) | `make test-scrape`, `make test-pipeline`, `make test-graph`; graph read/auth/MCP: `make test-graph-serve` |
| Graph read smoke (Docker) | `make test-graph-read-smoke` — no scrape/NATS |
| Graph version (ingest paths) | `./scripts/release/bump-graph-version.sh patch` → updates [versions.env](../versions.env) |
| Verify bump when required | `make check-graph-version` |
| Commit + push | See [AGENTS.md](../AGENTS.md) |

Ingest-affecting paths: `scrape/harvest/internal/sources/`, `pipeline/ned/internal/sources/`, `graph/ingest/internal/sources/`, `pkg/harvest/`, `pkg/commit/`, `docs/schemas/`.

---

## Markdown links (GitHub)

GitHub shows a folder icon only when the link target is a directory **and** the URL ends with `/`.

| Target | Link form |
|--------|-----------|
| Directory | `[label](../path/to/dir/)` — trailing `/` required |
| Layer README | `[Scrape](../scrape/README.md)` |
| File | `[compose.yml](../deploy/scrape/compose.yml)` — no trailing `/` |

Lint: `./scripts/housekeeping/lint-markdown-dir-links.sh`

---

## Go style (Google)

Follow the [Google Go Style Guide](https://google.github.io/styleguide/go/guide) and [best practices](https://google.github.io/styleguide/go/best-practices):

| Principle | In this repo |
|-----------|----------------|
| **Clarity** | Names and structure explain what and why; comments for rationale, not restating code |
| **Simplicity** | Prefer standard library over extra abstractions |
| **Consistency** | `gofmt`, **MixedCaps**, match neighboring files |
| **Package names** | Short, no `v1` in directory names; version only in `schema_version` |
| **Avoid repetition** | Do not repeat package name in exported symbols (`harvest.Envelope`, not `harvest.HarvestEnvelope`) |
| **Errors** | Wrap with `%w`; stable prefixes (`harvest:`, `commit:`); no `panic` in libraries; log once at `usecase` or `cmd` |
| **Tests** | Table-driven where useful; `testdata/` next to the package; Neo4j integration: build tag `integration` |

`pkg/` layout: one module ([pkg/go.mod](../pkg/go.mod)); scrape-only helpers under [scrape/pkg/](../scrape/pkg/).

---

## Naming

- Compose services / binaries: **`snake_case`** (`scrape_worker`, `pipeline_worker`, `ingest_worker`)
- NATS durable consumers: **`snake_case`** (`pipeline_worker`, `ingest_worker`)
- Go module path: `github.com/butbeautifulv/veil/...` (unchanged)

---

## Logging and lifecycle

- **`log/slog`** with structured attributes
- Long-running binaries: **`errgroup`** + cancel on **SIGINT/SIGTERM**
- Explicit **timeouts** on HTTP clients

---

## Configuration

- Environment variables with sensible defaults
- Document new env vars in [docs/threatintel-runtime.md](threatintel-runtime.md)

---

## License

**MIT** — [LICENSE](../LICENSE), [CONTRIBUTING.md](../CONTRIBUTING.md)
