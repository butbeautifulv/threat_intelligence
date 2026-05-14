# Deploy and scale (v0.3.0)

Stack version: see [VERSION](../VERSION). Graph pack release tag on GitHub: **`v0.3.0-graph-pack`** (asset `threat-intel-graph-v0.3.0.zip`).

## HTTP API behind nginx (load balancing)

Use profile **`deploy`** with [docker-compose.deploy.yml](../docker-compose.deploy.yml):

```bash
docker compose -f docker-compose.yml -f docker-compose.deploy.yml --profile deploy up --build -d
```

- Public entry: **`http://localhost:${LB_HTTP_PORT:-8888}`** (nginx).
- Upstream uses Docker embedded DNS (**`127.0.0.11`**) with a **variable** in **`proxy_pass`** so scaled **`api`** replicas are re-resolved (plain nginx OSS has no **`resolve`** on `server` in `upstream`).

Scale API replicas (stateless Bolt readers; all hit the same Neo4j):

```bash
docker compose -f docker-compose.yml -f docker-compose.deploy.yml --profile deploy up -d --scale api=3
```

Hardening in [docker/nginx/nginx.conf](../docker/nginx/nginx.conf): **`server_tokens off`**, security headers, **`limit_req`**, **`limit_conn`**, body size and proxy timeouts.

## Scale `ingest-worker` (NATS → Neo4j)

With **`--profile scrape`**, run several workers with the **same** `NATS_DURABLE` (default `ingest-worker`): JetStream shares messages across competing pull subscribers.

```bash
docker compose --profile scrape up -d --scale ingest-worker=3
```

Do **not** scale the same **scraper** service (e.g. `vuln`) without a deliberate partitioning strategy: duplicate scrapers repeat the same work unless feeds are sharded externally.

## Multistage Go images and BuildKit cache

Dockerfiles use a **build** stage plus minimal runtime image. Enable BuildKit for faster rebuilds:

```bash
DOCKER_BUILDKIT=1 docker compose build api
```

`go mod download` / `go build` use **`--mount=type=cache`** where the Dockerfile was updated (see `docker/*.Dockerfile`).

## GitHub graph pack release

1. Produce `data/neo4j_user_export/graph.cypher` (running Neo4j + [scripts/export-graph-cypher.sh](../scripts/export-graph-cypher.sh)).
2. `GRAPH_PACK_VERSION=v0.3.0 ./scripts/build-graph-pack.sh`
3. Create GitHub **Release** tag **`v0.3.0-graph-pack`** and attach **`data/neo4j_user_export/releases/threat-intel-graph-v0.3.0.zip`** so [docker/graph-bootstrap.sh](../docker/graph-bootstrap.sh) default URL resolves.

Example with GitHub CLI (needs `gh auth login`):

```bash
gh release create v0.3.0-graph-pack \
  --title "Graph pack v0.3.0" \
  --notes "Neo4j bootstrap ZIP (manifest + graph.cypher)." \
  data/neo4j_user_export/releases/threat-intel-graph-v0.3.0.zip
```

Repository app tag **`v0.3.0`** tracks the codebase / compose / docs line used for this release wave.
