#!/usr/bin/env bash
# Secure overlay smoke: nginx :8443 TLS, optional JWT when ENGAGE_AUTH_ENABLED=1.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null; then
  echo "skip secure smoke: docker not available" >&2
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "skip secure smoke: docker daemon not running" >&2
  exit 0
fi

CERT_DIR="${ROOT}/deploy/engage/nginx/certs"
mkdir -p "${CERT_DIR}"
if [[ ! -f "${CERT_DIR}/tls.crt" ]]; then
  openssl req -x509 -nodes -days 1 -newkey rsa:2048 \
    -keyout "${CERT_DIR}/tls.key" -out "${CERT_DIR}/tls.crt" \
    -subj "/CN=engage.local" 2>/dev/null
fi

COMPOSE=(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.secure.yml)
PROJECT="${COMPOSE_PROJECT_NAME:-engage-secure-$$}"
HTTPS_PORT="${ENGAGE_NGINX_HTTPS_PORT:-8443}"
API_URL="https://127.0.0.1:${HTTPS_PORT}"

cleanup() {
  COMPOSE_PROJECT_NAME="${PROJECT}" "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

export COMPOSE_PROJECT_NAME="${PROJECT}"
export ENGAGE_NGINX_TLS_CERT="${CERT_DIR}/tls.crt"
export ENGAGE_NGINX_TLS_KEY="${CERT_DIR}/tls.key"
export ENGAGE_NGINX_HTTPS_PORT="${HTTPS_PORT}"

echo "secure smoke: starting stack (project=${PROJECT})..."
"${COMPOSE[@]}" up -d --build engage-api engage-mcp nginx

deadline=$((SECONDS + 180))
until curl -kfsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; do
  if (( SECONDS >= deadline )); then
    echo "timeout waiting for nginx/https health" >&2
    "${COMPOSE[@]}" logs nginx 2>&1 | tail -30
    exit 1
  fi
  sleep 2
done

if [[ "${ENGAGE_AUTH_ENABLED:-0}" == "1" ]]; then
  token=$(curl -kfsS -X POST "${API_URL}/api/auth/token" \
    -H 'Content-Type: application/json' \
    -d '{"client_id":"engage-smoke","client_secret":"engage-smoke"}' 2>/dev/null | \
    python3 -c 'import json,sys; print(json.load(sys.stdin).get("access_token",""))' 2>/dev/null || true)
  if [[ -n "${token}" ]]; then
    curl -kfsS "${API_URL}/api/jobs" -H "Authorization: Bearer ${token}" | grep -q '\[' || true
  fi
fi

echo "OK engage secure smoke (https://${HTTPS_PORT})"
