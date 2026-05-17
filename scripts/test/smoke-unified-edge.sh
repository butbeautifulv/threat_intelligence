#!/usr/bin/env bash
# Minimal graph read + engage MCP HTTP behind unified platform nginx (P12i).
# Requires deploy/platform/compose.edge.yml (P12b).
set -euo pipefail
# shellcheck source=lib/smoke.sh
source "$(dirname "$0")/lib/smoke.sh"
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

EDGE_OVERLAY="deploy/platform/compose.edge.yml"
if [[ ! -f "${EDGE_OVERLAY}" ]]; then
  echo "[unified-edge-smoke] SKIP: missing ${EDGE_OVERLAY} (merge platform/p12b-veil-nginx-edge)" >&2
  exit 0
fi

smoke_skip_no_docker

export COMPOSE_FILES="-f deploy/knowledge/compose.yml -f deploy/knowledge/compose.graph-read.yml -f deploy/engage/compose.yml -f ${EDGE_OVERLAY} -f deploy/platform/compose.edge-hostless.yml"
export COMPOSE_PROFILES="${COMPOSE_PROFILES:-mcp}"
export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
# Avoid inheriting a fixed project name from ad-hoc debug runs.
unset COMPOSE_PROJECT_NAME
export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-unified-edge-smoke-$$}"
export VEIL_EDGE_HTTPS_PORT="${VEIL_EDGE_HTTPS_PORT:-8443}"
export SMOKE_UNIFIED_EDGE_WAIT_SEC="${SMOKE_UNIFIED_EDGE_WAIT_SEC:-300}"

CERT_DIR="${VEIL_ROOT}/deploy/platform/nginx/certs"
mkdir -p "${CERT_DIR}"
if [[ ! -f "${CERT_DIR}/tls.crt" ]]; then
  openssl req -x509 -nodes -days 1 -newkey rsa:2048 \
    -keyout "${CERT_DIR}/tls.key" -out "${CERT_DIR}/tls.crt" \
    -subj "/CN=veil.local" 2>/dev/null
fi
export VEIL_EDGE_TLS_CERT="${VEIL_EDGE_TLS_CERT:-${CERT_DIR}/tls.crt}"
export VEIL_EDGE_TLS_KEY="${VEIL_EDGE_TLS_KEY:-${CERT_DIR}/tls.key}"

EDGE_URL="https://127.0.0.1:${VEIL_EDGE_HTTPS_PORT}"
CURL_TLS=(curl -skS)

log() { printf '[unified-edge-smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; compose ps 2>/dev/null || true; exit 1; }

diag_logs() {
  for svc in veil-edge api mcp engage-api engage-mcp neo4j graph-bootstrap; do
    log "--- logs ${svc} (tail 40) ---"
    compose logs --tail=40 "${svc}" 2>/dev/null || true
  done
}

cleanup() {
  compose down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

BUILD_FLAG=()
if [[ "${SMOKE_UNIFIED_EDGE_BUILD:-1}" == "1" ]]; then
  BUILD_FLAG=(--build)
fi

log "starting minimal unified-edge stack (project=${COMPOSE_PROJECT_NAME}, port=${VEIL_EDGE_HTTPS_PORT})..."
compose up -d "${BUILD_FLAG[@]}" \
  neo4j graph-bootstrap api mcp engage-api engage-mcp veil-edge

wait_deadline=$((SECONDS + SMOKE_UNIFIED_EDGE_WAIT_SEC))
edge_ready=0
while (( SECONDS < wait_deadline )); do
  if "${CURL_TLS[@]}" -fsS "${EDGE_URL}/health" 2>/dev/null | grep -q '"ok"'; then
    edge_ready=1
    log "veil-edge /health ready"
    break
  fi
  sleep 3
done
if [[ "${edge_ready}" -ne 1 ]]; then
  diag_logs
  fail "timeout waiting for ${EDGE_URL}/health (${SMOKE_UNIFIED_EDGE_WAIT_SEC}s)"
fi

code="$("${CURL_TLS[@]}" -o /tmp/unified-categories.json -w '%{http_code}' "${EDGE_URL}/v1/categories" || echo 000)"
if [[ "${code}" != "200" ]]; then
  fail "GET /v1/categories => HTTP ${code}"
fi
log "OK GET /v1/categories (${code})"

engage_code="$("${CURL_TLS[@]}" -o /tmp/unified-engage-health.json -w '%{http_code}' "${EDGE_URL}/api/tools" || echo 000)"
if [[ "${engage_code}" != "200" ]]; then
  fail "engage-api via edge GET /api/tools => HTTP ${engage_code}"
fi
log "OK engage-api via edge GET /api/tools (${engage_code})"

mcp_initialize() {
  local url="$1" label="$2" expect="$3"
  local body='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"unified-edge-smoke","version":"0"}}}'
  local resp
  resp="$("${CURL_TLS[@]}" -X POST "${url}" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    -d "${body}" 2>/dev/null || true)"
  if ! echo "${resp}" | grep -q "${expect}"; then
    fail "MCP initialize ${label} at ${url} missing ${expect}: ${resp:0:400}"
  fi
  log "OK MCP initialize ${label}"
}

mcp_initialize "${EDGE_URL}/mcp/graph" "graph" '"protocolVersion"'
mcp_initialize "${EDGE_URL}/mcp/engage" "engage" 'veil-engage'

log "OK unified-edge smoke"
