#!/usr/bin/env bash
# CI smoke: minimal Veil stack + engage veil-stack overlay — tool run -> ingest -> veil-api engage search.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

if ! command -v docker >/dev/null 2>&1; then
  echo "SKIP: docker not available"
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "SKIP: docker daemon not running"
  exit 0
fi

export COMPOSE_FILES="${VEIL_COMPOSE_FILES} -f deploy/engage/compose.yml -f deploy/engage/compose.veil-stack.yml"
export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
export SMOKE_VEIL_API_WAIT_SEC="${SMOKE_VEIL_API_WAIT_SEC:-300}"
export SMOKE_VEIL_ENGAGE_WAIT_SEC="${SMOKE_VEIL_ENGAGE_WAIT_SEC:-180}"
export ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"
export API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"
export SMOKE_ENGAGE_HOST="${SMOKE_ENGAGE_HOST:-example.com}"

PROJECT="${COMPOSE_PROJECT_NAME:-engage-veil-ci-$$}"
export COMPOSE_PROJECT_NAME="${PROJECT}"

log() { printf '[veil-engage-ci] %s\n' "$*"; }

diag_logs() {
  compose ps 2>/dev/null || true
  for svc in engage-api api nats neo4j ingest_worker engage-events-worker; do
    log "--- logs ${svc} (tail 80) ---"
    compose logs --tail=80 "${svc}" 2>/dev/null || true
  done
}

fail() {
  log "FAIL: $*"
  diag_logs
  exit 1
}

cleanup() {
  compose down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

BUILD_FLAG=()
if [[ "${SMOKE_VEIL_STACK_BUILD:-1}" == "1" ]]; then
  BUILD_FLAG=(--build)
fi

log "starting minimal veil+engage stack (project=${PROJECT}, GRAPH_PACK_SKIP=${GRAPH_PACK_SKIP})..."
compose up -d "${BUILD_FLAG[@]}" \
  nats neo4j ingest_worker api engage-api engage-events-worker

api_wait_deadline=$((SECONDS + SMOKE_VEIL_API_WAIT_SEC))
until curl -sf "${ENGAGE_URL}/health" >/dev/null 2>&1; do
  if (( SECONDS >= api_wait_deadline )); then
    fail "timeout waiting for engage-api (${SMOKE_VEIL_API_WAIT_SEC}s)"
  fi
  sleep 2
done

veil_deadline=$((SECONDS + 120))
until curl -sf "${API_URL}/health" >/dev/null 2>&1; do
  if (( SECONDS >= veil_deadline )); then
    fail "timeout waiting for veil-api"
  fi
  sleep 2
done
log "engage-api and veil-api healthy"

log "POST engage tool run (httpx_probe -> ${SMOKE_ENGAGE_HOST}; audit event even if binary missing)"
code=$(curl -sS -o /tmp/engage-tool-out.json -w '%{http_code}' -X POST "${ENGAGE_URL}/api/tools/httpx_probe" \
  -H 'Content-Type: application/json' \
  -d "{\"target\":\"https://${SMOKE_ENGAGE_HOST}\"}" 2>/dev/null || echo 000)
if [[ "${code}" != "200" ]]; then
  fail "engage tool POST HTTP ${code}: $(cat /tmp/engage-tool-out.json 2>/dev/null || true)"
fi

poll_deadline=$((SECONDS + SMOKE_VEIL_ENGAGE_WAIT_SEC))
found=0
while (( SECONDS < poll_deadline )); do
  resp=$(curl -sf "${API_URL}/v1/categories/engage/search?q=${SMOKE_ENGAGE_HOST}&limit=10" 2>/dev/null || echo '{}')
  if echo "${resp}" | grep -qE 'EngageToolRun|EngageFinding|"total"'; then
    if command -v jq >/dev/null 2>&1; then
      count=$(echo "${resp}" | jq -r '.total // (.items | length) // 0' 2>/dev/null || echo 0)
      if [[ "${count}" =~ ^[0-9]+$ ]] && [[ "${count}" -ge 1 ]]; then
        found=1
        log "engage search hits: ${count}"
        break
      fi
    elif echo "${resp}" | grep -qi 'engagetoolrun\|engagefinding'; then
      found=1
      log "engage search returned engage nodes"
      break
    fi
  fi
  sleep 3
done

if [[ "${found}" -ne 1 ]]; then
  log "last search response: ${resp}"
  fail "expected engage category search for q=${SMOKE_ENGAGE_HOST} within ${SMOKE_VEIL_ENGAGE_WAIT_SEC}s"
fi

log "OK veil-engage stack CI smoke"
