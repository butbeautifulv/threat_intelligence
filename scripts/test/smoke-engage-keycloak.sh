#!/usr/bin/env bash
# Keycloak overlay smoke: AUTH_ENABLED requires JWT on API routes; health stays open.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null; then
  echo "skip keycloak smoke: docker not available" >&2
  exit 0
fi

COMPOSE=(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.keycloak.yml)
PROJECT="${COMPOSE_PROJECT_NAME:-engage-kc-$$}"
API_URL="http://127.0.0.1:8890"
KC_URL="http://127.0.0.1:8080"

cleanup() {
  COMPOSE_PROJECT_NAME="${PROJECT}" "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

export COMPOSE_PROJECT_NAME="${PROJECT}"
"${COMPOSE[@]}" up -d --build keycloak engage-api

deadline=$((SECONDS + 240))
until curl -fsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; do
  if (( SECONDS >= deadline )); then
    echo "timeout waiting for engage-api" >&2
    exit 1
  fi
  sleep 3
done

kc_deadline=$((SECONDS + 120))
until curl -fsS "${KC_URL}/realms/master" >/dev/null 2>&1; do
  if (( SECONDS >= kc_deadline )); then
    echo "timeout waiting for keycloak" >&2
    exit 1
  fi
  sleep 2
done

code=$(curl -sS -o /dev/null -w '%{http_code}' "${API_URL}/api/tools" || echo "000")
if [[ "${code}" != "401" ]]; then
  echo "FAIL: GET /api/tools without token expected 401, got ${code}" >&2
  exit 1
fi
echo "OK unauthenticated /api/tools returns 401"

token=$(curl -fsS -X POST "${KC_URL}/realms/master/protocol/openid-connect/token" \
  -d "client_id=admin-cli" \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=admin" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("access_token",""))' 2>/dev/null || true)
if [[ -z "${token}" ]]; then
  echo "FAIL: could not obtain Keycloak access token" >&2
  exit 1
fi

code=$(curl -sS -o /dev/null -w '%{http_code}' -H "Authorization: Bearer ${token}" "${API_URL}/api/tools" || echo "000")
if [[ "${code}" != "200" ]]; then
  echo "FAIL: GET /api/tools with token expected 200, got ${code}" >&2
  exit 1
fi
echo "OK authenticated /api/tools returns 200"

echo "OK engage keycloak smoke"
