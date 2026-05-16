#!/usr/bin/env bash
# Enable catalog tools when binaries exist on PATH (wrapper for enable-catalog-by-category.sh).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT="${ROOT}/scripts/engage/enable-catalog-by-category.sh"
CATEGORIES="${*:-network web osint cloud binary}"
exec "${SCRIPT}" ${CATEGORIES}
