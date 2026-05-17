#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

python3 "${ROOT}/scripts/engage/extract-decision-tables.py"

cd "${ROOT}/pkg"
env -u GOWORK go test ./decision/... -run TestEffectivenessParityWithLegacy -count=1

echo "OK decision engine parity"
