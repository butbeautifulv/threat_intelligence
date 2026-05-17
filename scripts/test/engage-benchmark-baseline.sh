#!/usr/bin/env bash
# P10f: record nmap_scan / nuclei_scan tool timings via engage-api (regression baseline, not KPI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

API_URL="${ENGAGE_API_URL:-${ENGAGE_URL:-http://127.0.0.1:8890}}"
RESULTS_JSON="${BENCHMARK_BASELINE_JSON:-${ROOT}/scripts/benchmark/results/baseline-tools.json}"
UTC_NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

log() { printf '[engage-benchmark-baseline] %s\n' "$*"; }

mkdir -p "$(dirname "${RESULTS_JSON}")"

write_json() {
  python3 -c 'import json,sys; json.dump(json.loads(sys.argv[1]), open(sys.argv[2],"w"), indent=2)' "$1" "${RESULTS_JSON}"
}

stub_skip() {
  local reason=$1
  log "STUB: ${reason}"
  write_json "$(printf '%s' "{\"skipped\":true,\"stub\":true,\"reason\":\"${reason}\",\"api_url\":\"${API_URL}\",\"generated_at\":\"${UTC_NOW}\",\"tools\":{}}")"
  exit 0
}

if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
  stub_skip "docker not available"
fi

if ! command -v curl >/dev/null 2>&1; then
  stub_skip "curl not available"
fi

if ! curl -sf "${API_URL}/health" >/dev/null 2>&1; then
  stub_skip "engage-api not reachable"
fi

elapsed() {
  local start=$1 end
  end=$(date +%s)
  echo $(( end - start ))
}

run_tool_sec() {
  local tool=$1 target=$2 params_json=$3
  local body sec t0
  body=$(python3 -c "import json; print(json.dumps({'target': '${target}', 'parameters': json.loads('''${params_json}''')}))")
  t0=$(date +%s)
  curl -sf -X POST "${API_URL}/api/tools/${tool}" \
    -H 'Content-Type: application/json' \
    -d "${body}" >/dev/null
  sec=$(elapsed "${t0}")
  echo "${sec}"
}

log "nmap_scan on 127.0.0.1 via ${API_URL}"
nmap_sec=$(run_tool_sec nmap_scan 127.0.0.1 '{"scan_type":"-sn","ports":"","additional_args":"-T4 --host-timeout 5s"}')

log "nuclei_scan on https://example.com via ${API_URL}"
nuclei_sec=$(run_tool_sec nuclei_scan https://example.com '{"templates":"","additional_args":"-silent -duc -timeout 5"}' || echo "-")

export ENG_BASELINE_OUT="${RESULTS_JSON}"
export ENG_BASELINE_GENERATED="${UTC_NOW}"
export ENG_BASELINE_API_URL="${API_URL}"
export ENG_BASELINE_NMAP="${nmap_sec}"
export ENG_BASELINE_NUCLEI="${nuclei_sec}"

python3 <<'PY'
import json
import os

path = os.environ["ENG_BASELINE_OUT"]
obj = {
    "skipped": False,
    "stub": False,
    "generated_at": os.environ["ENG_BASELINE_GENERATED"],
    "api_url": os.environ["ENG_BASELINE_API_URL"],
    "tools": {
        "nmap_scan": {"target": "127.0.0.1", "seconds": os.environ["ENG_BASELINE_NMAP"]},
        "nuclei_scan": {
            "target": "https://example.com",
            "seconds": os.environ["ENG_BASELINE_NUCLEI"],
        },
    },
}
with open(path, "w", encoding="utf-8") as f:
    json.dump(obj, f, indent=2)
PY

log "wrote ${RESULTS_JSON}"
log "baseline complete (regression tracking only — not a KPI gate)"
