#!/usr/bin/env bash
# Fast-rich graph pack profile (~25 min): all 7 sources, minimal NVD, no Atomic/MSF bulk.
# Usage: ./scripts/graph-pack/profile-fast-rich.sh [--no-down]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

log() { printf '[graph-pack-fast-rich] %s\n' "$*"; }

if [[ "${1:-}" != "--no-down" ]]; then
  log "stopping stack and removing volumes..."
  compose down -v --remove-orphans 2>/dev/null || true
fi

source_profile fast-rich

log "starting full stack (sources=${SCRAPE_SOURCES}, NVD_MAX_PAGES=${NVD_MAX_PAGES})..."
exec "${VEIL_ROOT}/scripts/ops/compose-up-full.sh"
