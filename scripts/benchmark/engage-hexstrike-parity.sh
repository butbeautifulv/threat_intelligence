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

RESULTS_JSON="${ROOT}/scripts/benchmark/results/latest.json"
UTC_NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

write_latest_json_py() {
  python3 -c 'import json,sys; json.dump(json.loads(sys.argv[1]), open(sys.argv[2],"w"), indent=2)' "$1" "${RESULTS_JSON}"
}

mkdir -p "$(dirname "${RESULTS_JSON}")"

log() { printf '[engage-benchmark] %s\n' "$*"; }

if ! command -v curl >/dev/null 2>&1; then
  log "SKIP: curl not available"
  write_latest_json_py "$(printf '%s' "{\"skipped\":true,\"reason\":\"curl not available\",\"api_url\":\"${API_URL}\",\"generated_at\":\"${UTC_NOW}\"}")"
  exit 0
fi

if ! curl -sf "${API_URL}/health" >/dev/null 2>&1; then
  log "SKIP: engage-api not reachable at ${API_URL}"
  write_latest_json_py "$(printf '%s' "{\"skipped\":true,\"reason\":\"api unreachable\",\"api_url\":\"${API_URL}\",\"generated_at\":\"${UTC_NOW}\"}")"
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

export ENG_BENCH_OUT="${RESULTS_JSON}"
export ENG_BENCH_GENERATED="${UTC_NOW}"
export ENG_BENCH_API_URL="${API_URL}"
export ENG_BENCH_TARGET="${TARGET}"
export ENG_BENCH_EXECUTE="${EXECUTE}"
export ENG_BENCH_RECON="${recon_sec}"
export ENG_BENCH_SMART="${smart_sec}"
export ENG_BENCH_ASSESS="${assessment_sec}"
export ENG_BENCH_TABLE="${table}"

python3 <<'PY'
import json
import os

path = os.environ["ENG_BENCH_OUT"]
obj = {
    "skipped": False,
    "generated_at": os.environ["ENG_BENCH_GENERATED"],
    "api_url": os.environ["ENG_BENCH_API_URL"],
    "target": os.environ["ENG_BENCH_TARGET"],
    "execute": os.environ["ENG_BENCH_EXECUTE"] == "1",
    "seconds": {
        "reconnaissance_workflow": os.environ["ENG_BENCH_RECON"],
        "smart_scan": os.environ["ENG_BENCH_SMART"],
        "assessment_report": os.environ["ENG_BENCH_ASSESS"],
    },
    "markdown_table": os.environ["ENG_BENCH_TABLE"],
}
with open(path, "w", encoding="utf-8") as f:
    json.dump(obj, f, indent=2)
PY

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

log "wrote ${RESULTS_JSON}"
log "benchmark complete"
