#!/usr/bin/env bash
# Smoke: engage tool run -> engage.events -> bridge -> ingest.engage.* -> Neo4j (graph-ingest profile)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

if ! command -v docker >/dev/null 2>&1; then
  echo "SKIP: docker not available"
  exit 0
fi

COMPOSE="docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml"
$COMPOSE --profile graph-ingest up -d --build nats neo4j engage-api engage-events-worker ingest_worker 2>/dev/null || {
  echo "SKIP: compose events profile unavailable"
  exit 0
}

trap '$COMPOSE --profile graph-ingest down -v 2>/dev/null || true' EXIT

for i in $(seq 1 60); do
  if curl -sf http://127.0.0.1:8890/health >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

curl -sf -X POST http://127.0.0.1:8890/api/tools/httpx_probe \
  -H 'Content-Type: application/json' \
  -d '{"target":"https://example.com"}' >/dev/null || true

sleep 8

if docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml exec -T nats \
  nats stream info INGEST 2>/dev/null | grep -q 'Messages'; then
  echo "INGEST stream present"
fi

msgs=$(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml exec -T nats \
  nats stream view INGEST --subject ingest.engage.tool_run --count 1 2>/dev/null | wc -l || echo 0)
if [[ "${msgs}" -lt 1 ]]; then
  echo "WARN: could not confirm ingest.engage.tool_run message (nats CLI)" >&2
fi

count=$($COMPOSE --profile graph-ingest exec -T neo4j \
  cypher-shell -u neo4j -p neo4jpassword \
  "MATCH (r:EngageToolRun) RETURN count(r) AS c" 2>/dev/null | tail -1 | tr -d '[:space:]' || echo 0)
if [[ "${count}" =~ ^[0-9]+$ ]] && [[ "${count}" -ge 1 ]]; then
  echo "EngageToolRun nodes in Neo4j: ${count}"
else
  echo "FAIL: expected EngageToolRun count >= 1 in Neo4j, got '${count}'" >&2
  exit 1
fi

echo "OK engage-events-pipeline smoke"
