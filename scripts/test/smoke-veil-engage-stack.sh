#!/usr/bin/env bash
# Smoke: full Veil stack with engage on shared NATS — tool run -> ingest -> veil-api engage search.
# Prerequisite: ./scripts/ops/compose-up-veil-engage.sh (or equivalent compose with compose.veil-stack.yml).
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

ENGAGE_URL="${ENGAGE_URL:-http://127.0.0.1:${ENGAGE_API_PORT:-8890}}"
API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}}"
HOST="${SMOKE_ENGAGE_HOST:-example.com}"
WAIT_SEC="${SMOKE_VEIL_ENGAGE_WAIT_SEC:-120}"

log() { printf '[veil-engage-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if ! command -v curl >/dev/null 2>&1; then
  log "SKIP: curl not available"
  exit 0
fi

if ! curl -sf "${ENGAGE_URL}/health" >/dev/null 2>&1; then
  log "SKIP: engage-api not reachable at ${ENGAGE_URL} (run compose-up-veil-engage.sh first)"
  exit 0
fi

if ! curl -sf "${API_URL}/health" >/dev/null 2>&1; then
  log "SKIP: veil-api not reachable at ${API_URL}"
  exit 0
fi

log "POST engage tool run (httpx_probe -> ${HOST})"
curl -sf -X POST "${ENGAGE_URL}/api/tools/httpx_probe" \
  -H 'Content-Type: application/json' \
  -d "{\"target\":\"https://${HOST}\"}" >/dev/null || \
  curl -sf -X POST "${ENGAGE_URL}/api/tools/nmap_scan" \
    -H 'Content-Type: application/json' \
    -d "{\"target\":\"${HOST}\"}" >/dev/null || true

deadline=$((SECONDS + WAIT_SEC))
found=0
search_tmp="${TMPDIR:-/tmp}/veil-engage-search-$$.json"
while (( SECONDS < deadline )); do
  http=$(curl -sS -o "${search_tmp}" -w '%{http_code}' \
    "${API_URL}/v1/categories/engage/search?q=${HOST}&limit=10" 2>/dev/null || echo 000)
  resp=$(cat "${search_tmp}" 2>/dev/null || true)
  if [[ "${http}" == "200" ]] && command -v jq >/dev/null 2>&1; then
    count=$(echo "${resp}" | jq -r '((.nodes // []) | length)' 2>/dev/null || echo 0)
    if [[ "${count}" =~ ^[0-9]+$ ]] && [[ "${count}" -ge 1 ]]; then
      found=1
      log "engage search hits: ${count}"
      break
    fi
  elif [[ "${http}" == "200" ]] && echo "${resp}" | grep -qi 'engagetoolrun\|engagefinding'; then
    found=1
    log "engage search returned engage nodes"
    break
  fi
  sleep 3
done
rm -f "${search_tmp}" 2>/dev/null || true

if [[ "${found}" -ne 1 ]]; then
  fail "expected engage category search for q=${HOST} to return results within ${WAIT_SEC}s"
fi

log "OK veil-engage stack smoke"
