#!/usr/bin/env bash
# Benchmark: HexStrike-style KPI timing for engage-api (regression tracking).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

API_URL="${ENGAGE_API_URL:-${ENGAGE_URL:-http://127.0.0.1:8890}}"
TARGET="${BENCHMARK_TARGET:-example.com}"
EXECUTE="${ENGAGE_BENCHMARK_EXECUTE:-1}"
MAX_SEC="${BENCHMARK_MAX_SEC:-1800}"
SMART_MAX_TOOLS="${BENCHMARK_SMART_MAX_TOOLS:-5}"

log() { printf '[engage-benchmark] %s\n' "$*"; }

if ! command -v curl >/dev/null 2>&1; then
  log "SKIP: curl not available"
  exit 0
fi

if ! curl -sf "${API_URL}/health" >/dev/null 2>&1; then
  log "SKIP: engage-api not reachable at ${API_URL}"
  exit 0
fi

elapsed() {
  local start=$1 end
  end=$(date +%s)
  echo $(( end - start ))
}

warn_if_slow() {
  local name=$1 sec=$2 limit=$3
  if [[ "${sec}" -gt "${limit}" ]]; then
    log "WARN: ${name} took ${sec}s (soft limit ${limit}s)"
  fi
}

recon_sec="-"
smart_sec="-"
assessment_sec="-"

if [[ "${EXECUTE}" == "1" ]]; then
  log "Step 1: bugbounty recon (execute=true) target=${TARGET}"
  t0=$(date +%s)
  curl -sf -X POST "${API_URL}/api/bugbounty/reconnaissance-workflow" \
    -H 'Content-Type: application/json' \
    -d "{\"domain\":\"${TARGET}\",\"execute\":true}" >/dev/null
  recon_sec=$(elapsed "${t0}")
  warn_if_slow "recon" "${recon_sec}" "${MAX_SEC}"

  log "Step 2: smart-scan (comprehensive) target=${TARGET}"
  t0=$(date +%s)
  curl -sf -X POST "${API_URL}/api/intelligence/smart-scan" \
    -H 'Content-Type: application/json' \
    -d "{\"target\":\"${TARGET}\",\"objective\":\"comprehensive\",\"max_tools\":${SMART_MAX_TOOLS}}" >/dev/null
  smart_sec=$(elapsed "${t0}")
  warn_if_slow "smart-scan" "${smart_sec}" "${MAX_SEC}"

  log "Step 3: assessment-report target=${TARGET}"
  t0=$(date +%s)
  curl -sf -X POST "${API_URL}/api/intelligence/assessment-report" \
    -H 'Content-Type: application/json' \
    -d "{\"target\":\"${TARGET}\",\"objective\":\"comprehensive\",\"max_tools\":${SMART_MAX_TOOLS}}" >/dev/null
  assessment_sec=$(elapsed "${t0}")
  warn_if_slow "assessment-report" "${assessment_sec}" 300
else
  log "ENGAGE_BENCHMARK_EXECUTE=0 — dry run (health only)"
fi

table="| Step | Endpoint | Seconds |
|------|----------|---------|
| Subdomain / recon | POST /api/bugbounty/reconnaissance-workflow | ${recon_sec} |
| Vuln scan | POST /api/intelligence/smart-scan | ${smart_sec} |
| Assessment report | POST /api/intelligence/assessment-report | ${assessment_sec} |"

printf '\n%s\n\n' "${table}"
printf 'target=%s api=%s execute=%s\n' "${TARGET}" "${API_URL}" "${EXECUTE}"

if [[ -n "${BENCHMARK_OUT:-}" ]]; then
  mkdir -p "$(dirname "${BENCHMARK_OUT}")"
  {
    echo "# Engage HexStrike parity benchmark"
    echo ""
    echo "- Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "- Target: ${TARGET}"
    echo "- API: ${API_URL}"
    echo ""
    echo "${table}"
  } >"${BENCHMARK_OUT}"
  log "wrote ${BENCHMARK_OUT}"
fi

log "benchmark complete"
