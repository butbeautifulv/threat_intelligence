#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

python3 "${ROOT}/scripts/engage/extract-decision-tables.py"

cd "${ROOT}/engage/serve"
env GOWORK="$(dirname "$(pwd)")/go.work" go test ./internal/usecase/intelligence/... -run TestEffectivenessParityWithLegacy -count=1

echo "OK decision engine parity"
