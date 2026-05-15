#!/usr/bin/env bash
# Start full Veil stack (scrape + pipeline + graph) with optional worker scaling.
# Usage (from repo root):
#   ./scripts/compose-up-full.sh
#   PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2 ./scripts/compose-up-full.sh
#   SCRAPE_WORKER_PARTITION=1 ./scripts/compose-up-full.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE="${COMPOSE:-docker compose}"
COMPOSE_FILES="${COMPOSE_FILES:--f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml}"
PIPELINE_SCALE="${PIPELINE_WORKER_SCALE:-1}"
INGEST_SCALE="${INGEST_WORKER_SCALE:-1}"
SCRAPE_PARTITION="${SCRAPE_WORKER_PARTITION:-0}"

compose() {
  # shellcheck disable=SC2086
  $COMPOSE $COMPOSE_FILES "$@"
}

EXTRA_FILES=()
if [[ "$SCRAPE_PARTITION" == "1" ]]; then
  EXTRA_FILES+=(-f deploy/compose.scale.yml)
  export COMPOSE_PROFILES="${COMPOSE_PROFILES:-scrape-partition}"
fi

scale_args=(--scale "pipeline_worker=${PIPELINE_SCALE}" --scale "ingest_worker=${INGEST_SCALE}")

log() { printf '[compose-up-full] %s\n' "$*"; }

log "bringing up infrastructure (crawl-db, nats, neo4j, graph-bootstrap)..."
# shellcheck disable=SC2086
$COMPOSE $COMPOSE_FILES "${EXTRA_FILES[@]}" up -d --build crawl-db nats neo4j graph-bootstrap

log "starting workers (pipeline_scale=${PIPELINE_SCALE}, ingest_scale=${INGEST_SCALE}, scrape_partition=${SCRAPE_PARTITION})..."
if [[ "$SCRAPE_PARTITION" == "1" ]]; then
  # shellcheck disable=SC2086
  $COMPOSE $COMPOSE_FILES "${EXTRA_FILES[@]}" up -d --build "${scale_args[@]}" \
    scrape_worker_fast scrape_worker_slow pipeline_worker ingest_worker
else
  # shellcheck disable=SC2086
  $COMPOSE $COMPOSE_FILES "${EXTRA_FILES[@]}" up -d --build "${scale_args[@]}" \
    scrape_worker pipeline_worker ingest_worker
fi

log "optional: api"
# shellcheck disable=SC2086
$COMPOSE $COMPOSE_FILES "${EXTRA_FILES[@]}" up -d api 2>/dev/null || true

log "done — check: docker compose ps"
