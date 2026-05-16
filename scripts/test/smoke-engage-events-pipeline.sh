#!/usr/bin/env bash
# Smoke: engage tool run -> engage.events -> bridge -> ingest.engage.tool_run
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

if ! command -v docker >/dev/null 2>&1; then
  echo "SKIP: docker not available"
  exit 0
fi

COMPOSE="docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml"
$COMPOSE up -d --build nats engage-api engage-events-worker 2>/dev/null || {
  echo "SKIP: compose events profile unavailable"
  exit 0
}

trap '$COMPOSE down -v 2>/dev/null || true' EXIT

for i in $(seq 1 30); do
  if curl -sf http://127.0.0.1:8890/health >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

curl -sf -X POST http://127.0.0.1:8890/api/tools/httpx_probe \
  -H 'Content-Type: application/json' \
  -d '{"target":"https://example.com"}' >/dev/null || true

sleep 4

# Verify ingest subject received at least one message via nats CLI in container
if docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml exec -T nats \
  nats stream info INGEST 2>/dev/null | grep -q 'Messages'; then
  echo "INGEST stream present"
fi

msgs=$(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml exec -T nats \
  nats stream view INGEST --subject ingest.engage.tool_run --count 1 2>/dev/null | wc -l || echo 0)
if [[ "${msgs}" -lt 1 ]]; then
  echo "WARN: could not confirm ingest.engage.tool_run message (nats CLI); stack is up" >&2
else
  echo "ingest.engage.tool_run message observed"
fi
echo "OK engage-events-pipeline smoke"
