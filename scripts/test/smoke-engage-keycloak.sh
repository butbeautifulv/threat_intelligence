#!/usr/bin/env bash
# Keycloak overlay smoke: stack starts with AUTH_ENABLED (token fetch best-effort).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null; then
  echo "skip keycloak smoke: docker not available" >&2
  exit 0
fi

COMPOSE=(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.keycloak.yml)
PROJECT="${COMPOSE_PROJECT_NAME:-engage-kc-$$}"

cleanup() {
  COMPOSE_PROJECT_NAME="${PROJECT}" "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

export COMPOSE_PROJECT_NAME="${PROJECT}"
"${COMPOSE[@]}" up -d --build keycloak engage-api

deadline=$((SECONDS + 240))
until curl -fsS "http://127.0.0.1:8890/health" 2>/dev/null | grep -q '"ok":true'; do
  if (( SECONDS >= deadline )); then
    echo "timeout waiting for engage-api" >&2
    exit 1
  fi
  sleep 3
done
echo "OK engage keycloak compose smoke (api healthy with AUTH_ENABLED)"
