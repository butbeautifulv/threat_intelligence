#!/usr/bin/env bash
# Runner profile smoke: sync tool runs via API (recon chain tools).
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
run_tool subfinder_scan example.com '{"additional_args":"-silent"}'
run_tool httpx_probe https://example.com '{"additional_args":"-silent -status-code"}'
run_tool httpx_tech_detect https://example.com '{"additional_args":"-silent -tech-detect -status-code"}' || run_tool httpx_probe https://example.com '{"additional_args":"-silent -tech-detect"}'
run_tool nuclei_scan https://example.com '{"templates":"","additional_args":"-silent -duc"}' || true
run_tool katana_crawl https://example.com '{"additional_args":"-silent -d 1"}' || run_tool katana_depth_scan https://example.com '{"additional_args":"-silent -d 1"}' || true
echo "OK engage runner profile smoke"
