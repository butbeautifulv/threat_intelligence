# Coding style (threat_intelligence)

Conventions for **scrapers**, **`ingest-worker`**, and small **Go services** in this repo. When in doubt, mirror [scrapers/ti](../scrapers/ti) and [scrapers/vuln](../scrapers/vuln).

**Cursor and other agents:** treat this file as the source of truth for code layout and ingest rules; the repo entry point is [AGENTS.md](../AGENTS.md).

---

## Layering

Keep a clear dependency direction (no cycles):

1. **`cmd/`** — wiring only: parse flags/env, build `config`, open connections, construct `usecase`, call `Run(ctx)`. No Cypher in `cmd`.
2. **`internal/domain/`** *(optional)* — plain types, enums, validation rules that do not depend on Neo4j or HTTP clients.
3. **`internal/repository/`** — **interfaces** (“ports”) that `usecase` calls: `UpsertX`, `ListY`, etc. Implementations live under **`storage/`** at module root.
4. **`internal/usecase/`** — orchestration: loops, limits, backoff, **`INGEST_MODE`** (`direct` vs `nats`), calling `repository` and optional JetStream publisher.
5. **`internal/feeds/`** (or `internal/connector/`) — outbound HTTP, GitHub API, OSV, zip download: return DTOs or raw bytes / maps for parsers.
6. **`storage/`** (package at module root, **not** under `internal/`) — Neo4j **implementations** of `repository` so another module (**`ingest-worker`**) can import the same writers without violating `internal` rules.

`internal/*` is private to that Go module; anything another module must call should be under a non-`internal` path (typically **`storage/`**).

---

## Logging and lifecycle

- Use **`log/slog`** with structured attributes (`slog.String("feed", …)`).
- Long-running binaries: **`errgroup`** + `context` cancel on **SIGINT/SIGTERM** (see `scrapers/vuln/cmd`).
- Prefer explicit **timeouts** on HTTP clients (`http.Client{Timeout: …}`).

---

## Errors

- Wrap with `%w` when crossing package boundaries; log once at a sensible boundary (`usecase` or `cmd`), not inside every helper.
- Do not silently ignore fetch errors without at least a debug/info log.

---

## Configuration

- Environment variables with sensible defaults; flags may override for local runs.
- Document new env vars in [scrapers/README.md](../scrapers/README.md) and [threatintel-runtime.md](threatintel-runtime.md) when behaviour is user-visible.

---

## Tests

- **Table-driven** unit tests for pure logic: parsing, normalization, idempotency keys, envelope validation ([pkg/ingestv1](../pkg/ingestv1/)).
- Neo4j integration tests are optional; use a build tag (e.g. `integration`) if added later.

---

## Ingest pipeline (NATS)

- **Envelope:** [pkg/ingestv1](../pkg/ingestv1/) — versioned JSON (`schema_version`, `source`, `kind`, `idempotency_key`, `payload`).
- **Producers:** `sbom`, `coderules`, `nuclei`, and **`ti`** in **`INGEST_MODE=nats`** publish via [scrapers/ingestpub](../scrapers/ingestpub/). For **`sbom`** OSV, set **`SBOM_CVE_LIST_FILE`** or **`SBOM_CVE_LIST_URL`** so CVE ids are not read from Neo4j in the scraper process.
- **Consumer:** [scrapers/ingest-worker](../scrapers/ingest-worker/README.md) applies the same **`storage/neo4j`** writers as **`direct`** mode.
- Default **`INGEST_MODE=direct`** keeps `docker compose run sbom` usable without NATS until you opt in.

---

## Git

After meaningful chunks of work: **commit** with a clear message and **push** the branch to avoid losing work.

## License

Project license: **MIT** — see [LICENSE](../LICENSE) in the repository root. Contributing implies acceptance of that license; see [CONTRIBUTING.md](../CONTRIBUTING.md).
