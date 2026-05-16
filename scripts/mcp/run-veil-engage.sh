#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}/engage/serve"
export ENGAGE_ENV="${ENGAGE_ENV:-local}"
export ENGAGE_CATALOG_PATH="${ENGAGE_CATALOG_PATH:-${ROOT}/engage/serve/catalog/tools.yaml}"
export AUTH_ENABLED="${AUTH_ENABLED:-0}"
exec env GOWORK="${ROOT}/engage/go.work" go run ./cmd/mcp
