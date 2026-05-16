#!/usr/bin/env bash
# Full Veil stack (scrape + pipeline + graph) plus engage-api and engage-events bridge on shared NATS.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

export COMPOSE_FILES="${VEIL_COMPOSE_FILES} -f deploy/engage/compose.yml -f deploy/engage/compose.veil-stack.yml"

log() { printf '[compose-up-veil-engage] %s\n' "$*"; }

"${VEIL_ROOT}/scripts/ops/compose-up-full.sh"

log "starting engage-api and engage-events-worker..."
compose up -d --build engage-api engage-events-worker

log "done — engage-api :8890, veil-api :8090, shared NATS/Neo4j"
log "smoke: ./scripts/test/smoke-veil-engage-stack.sh"
