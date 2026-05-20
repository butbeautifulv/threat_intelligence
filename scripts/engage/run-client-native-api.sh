#!/usr/bin/env bash
# Run engage-api on the host with client-native execution (host PATH, no docker runner).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}/engage/serve"
export ENGAGE_ENV="${ENGAGE_ENV:-local}"
export ENGAGE_EXECUTION_PROFILE="${ENGAGE_EXECUTION_PROFILE:-client-native}"
export ENGAGE_RUNNER_MODE="${ENGAGE_RUNNER_MODE:-local}"
unset ENGAGE_RUNNER_CONTAINER || true
export ENGAGE_API_LISTEN="${ENGAGE_API_LISTEN:-:8890}"
export ENGAGE_CATALOG_PATH="${ENGAGE_CATALOG_PATH:-${ROOT}/engage/serve/catalog/tools.yaml}"
export AUTH_ENABLED="${AUTH_ENABLED:-0}"
exec env GOWORK="${ROOT}/engage/go.work" go run ./cmd/api
