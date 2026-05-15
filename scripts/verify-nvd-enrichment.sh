#!/usr/bin/env bash
# Check that NVD ingest created CWE/CPE relationships (post-scrape or post-pack import).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE_FILES="${COMPOSE_FILES:--f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml}"

cd "$ROOT"
# shellcheck disable=SC2086
docker compose $COMPOSE_FILES exec -T neo4j cypher-shell -u neo4j -p neo4jpassword \
  "MATCH (v:Vulnerability)-[:HAS_CWE]->() RETURN count(*) AS has_cwe;
   MATCH (v:Vulnerability)-[:AFFECTS]->(:CPE) RETURN count(*) AS affects;
   MATCH (n:CPE) RETURN count(n) AS cpe_nodes;"
