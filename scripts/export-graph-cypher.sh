#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_HOST="${NEO4J_EXPORT_HOST_DIR:-${ROOT}/data/neo4j_user_export}"
mkdir -p "${OUT_HOST}"

COMPOSE="${COMPOSE:-docker compose}"
COMPOSE_FILES="${COMPOSE_FILES:--f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"

# Writes under Neo4j import dir (mounted as ./data/neo4j_user_export → /var/lib/neo4j/import/user_export by default)
OUT_REL="user_export/graph.cypher"

compose() {
  # shellcheck disable=SC2086
  (cd "${ROOT}" && $COMPOSE $COMPOSE_FILES "$@")
}

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" -d neo4j \
    "CALL apoc.export.cypher.all('${OUT_REL}', {format: 'cypher-shell'}) YIELD file, batches, source, format RETURN file, batches, source, format;"
else
  echo "docker compose not available; run manually against Neo4j:" >&2
  echo "  cypher-shell ... \"CALL apoc.export.cypher.all('${OUT_REL}', {format: 'cypher-shell'})\"" >&2
  exit 1
fi

echo "Export written to ${OUT_HOST}/graph.cypher"
