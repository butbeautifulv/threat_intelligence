# Coding style (threat_intelligence)

Conventions for **scrapers**, **workers**, and small **Go services** in this repo. When in doubt, mirror [scrapers/ti](https://github.com/butbeautifulv/threat_intelligence/tree/main/scrapers/ti) and [scrapers/vuln](https://github.com/butbeautifulv/threat_intelligence/tree/main/scrapers/vuln).

## Layering

Keep a clear dependency direction (no cycles):

1. **`cmd/`** — wiring only: parse flags/env, build `config`, open connections, construct `usecase`, call `Run(ctx)`. No HTTP details beyond what the binary owns; no Cypher.
2. **`internal/domain/`** — plain types, enums, validation rules that do not depend on Neo4j or HTTP clients.
3. **`internal/repository/`** — **interfaces** (“ports”) that `usecase` calls: `UpsertX`, `ListY`, etc. Implementations live under `storage/`.
4. **`internal/usecase/`** — orchestration: loops, limits, backoff, `INGEST_MODE` (`direct` vs `nats`), calling `repository` and optional `publisher`.
5. **`internal/feeds/`** (or `internal/connector/`) — outbound HTTP, GitHub API, OSV, zip download: return domain DTOs or `[]byte` / `map` for parsers.
6. **`storage/`** (package at module root, **not** under `internal/`) — Neo4j (or other) **implementations** of `repository` so a separate binary (e.g. `ingest-worker`) can import the same writer without violating `internal` rules.

`internal/*` is private to that Go module; anything another module must call should be under a non-`internal` path (typically `storage/`).

## Logging and lifecycle

- Use **`log/slog`** with structured attributes (`slog.String("feed", …)`).
- Long-running binaries: **`errgroup`** + `context` cancel on **SIGINT/SIGTERM** (see `vuln/cmd`).
- Prefer explicit **timeouts** on HTTP clients (`http.Client{Timeout: …}`).

## Errors

- Wrap with `%w` when crossing package boundaries; log once at a sensible boundary (usecase or cmd), not inside every helper.
- Do not silently ignore fetch errors without at least a debug/info log.

## Configuration

- Environment variables with sensible defaults; flags may override for local runs.
- Document new env vars in [scrapers/README.md](../scrapers/README.md) and [docs/threatintel-runtime.md](threatintel-runtime.md) when behavior is user-visible.

## Tests

- **Table-driven** unit tests for pure logic: parsing, normalization, idempotency keys, envelope validation (`pkg/ingestv1`).
- Neo4j integration tests are optional; use a build tag (e.g. `integration`) if added later.

## Ingest pipeline (NATS)

- **Envelope:** `pkg/ingestv1` — versioned JSON (`schema_version`, `source`, `kind`, `idempotency_key`, `payload`).
- **Producers** (scrapers) in `INGEST_MODE=nats` publish to JetStream; **consumer** (`ingest-worker`) applies the same `storage` writers as `direct` mode.
- Default remains **`direct`** so `docker compose run sbom` works without NATS until explicitly enabled.

## Git

After meaningful chunks of work: **commit** with a clear message and **push** the branch to avoid losing work.
