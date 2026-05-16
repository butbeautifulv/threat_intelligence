#!/usr/bin/env bash
# Runner profile smoke: sync tool runs via API (nmap, nuclei, httpx) when stack is up.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null; then
  echo "skip runner profile smoke: docker not available" >&2
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "skip runner profile smoke: docker daemon not running" >&2
  exit 0
fi

API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"
if ! curl -fsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; then
  echo "skip runner profile smoke: engage-api not reachable at ${API_URL}" >&2
  exit 0
fi

run_tool() {
  local tool=$1 target=$2 extra=${3:-'{}'}
  local body
  body=$(python3 -c "import json; print(json.dumps({'target': '${target}', 'parameters': json.loads('''${extra}''')}))")
  curl -fsS -X POST "${API_URL}/api/tools/${tool}" \
    -H 'Content-Type: application/json' \
    -d "${body}" >/dev/null
  echo "  ok ${tool}"
}

echo "runner profile smoke: sync tools via ${API_URL}"
run_tool nmap_scan 127.0.0.1 '{"scan_type":"-sn","ports":"","additional_args":"-T4"}'
run_tool httpx_probe https://example.com '{"additional_args":"-silent -status-code"}'
run_tool nuclei_scan https://example.com '{"templates":"","additional_args":"-silent -duc"}' || true
echo "OK engage runner profile smoke"
