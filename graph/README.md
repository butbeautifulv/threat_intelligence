# Graph layer

Architecture rules: [docs/coding-style.md](../docs/coding-style.md).

Consumes `ingest.>`, MERGE into Neo4j; HTTP API and MCP read Bolt.

| Module | Path | Role |
|--------|------|------|
| **ingest** | [ingest/](ingest/) | JetStream pull consumer (`ingest_worker`) → Neo4j |
| **serve** | [serve/](serve/) | HTTP API (`api`) and stdio MCP (`mcp`) |
| **connector** | [connector/](connector/) | Shared Bolt driver and categorical queries |

- **Wire types:** [pkg/commit](../pkg/commit/)
- **Build:** `make test-graph` or:

```bash
cd graph/ingest && go build -o bin/ingest_worker ./cmd/ingest_worker
cd graph/serve && go build -o bin/api ./cmd/api && go build -o bin/mcp ./cmd/mcp
```

- **Deploy:** [deploy/graph/compose.yml](../deploy/graph/compose.yml)

## ingest layout

```
ingest/
  cmd/ingest_worker/     # thin main
  internal/
    components/          # DI: Neo4j writers + domain appliers
    ingest/              # NATS pull loop
    sources/{ti,vuln,lola,ds}/
    appsec/{sbom,coderules,nuclei}/
```

## serve layout

```
serve/
  cmd/{api,mcp}/
  internal/
    components/ usecase/ transport/ storage/
```
