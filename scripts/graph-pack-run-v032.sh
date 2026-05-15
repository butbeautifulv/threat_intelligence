#!/usr/bin/env bash
# Fast-rich graph pack v0.3.2 profile (~25 min): all 7 sources, minimal NVD, no Atomic/MSF bulk.
# Usage (from repo root):
#   ./scripts/graph-pack-run-v032.sh
#   ./scripts/graph-pack-run-v032.sh --no-down   # skip volume wipe if stack already clean
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

log() { printf '[graph-pack-v032] %s\n' "$*"; }

if [[ "${1:-}" != "--no-down" ]]; then
  log "stopping stack and removing volumes..."
  docker compose -f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml \
    down -v --remove-orphans 2>/dev/null || true
fi

# Fast-rich profile (do not inherit smoke limits from the shell).
export GRAPH_PACK_SKIP=1
export SCRAPE_FORCE_REFETCH=1
export SCRAPE_SOURCES=ds,vuln,lola,ti,sbom,coderules,nuclei

export NVD_MAX_PAGES=1
export VULN_METASPLOIT_MAX_RB=0
export VULN_EXPLOITDB_MAX_ROWS=5000

export DS_MAX_ATOMIC=0
export DS_MAX_SIGMA=120
export DS_MAX_YARA=80
export DS_MAX_CALDERA=40

export SBOM_SOURCES=osv,ghsa
export SBOM_MAX_CVES=80
export SBOM_MAX_GHSA=50

export CODERULES_MAX_SEMGREP=40
export CODERULES_MAX_CODEQL=30
export NUCLEI_MAX=60

export LOLA_MITRE_MAX_TECHNIQUES=2000
export LOFTS_SKIP_ON_ERROR=true

log "starting full stack (sources=${SCRAPE_SOURCES}, NVD_MAX_PAGES=${NVD_MAX_PAGES})..."
exec "$ROOT/scripts/compose-up-full.sh"
