#!/usr/bin/env bash
# Aggressive HTTP abuse harness against a local Engage victim API only.
# Requires: curl. Optional: jq (not required for pass/fail).
# Env: ENGAGE_VICTIM_URL (default http://127.0.0.1:8891)
set -euo pipefail

V="${ENGAGE_VICTIM_URL:-http://127.0.0.1:8891}"
V="${V%/}"
AUTH_HDR=()
if [[ -n "${ENGAGE_VICTIM_TOKEN:-}" ]]; then
  AUTH_HDR=(-H "Authorization: Bearer ${ENGAGE_VICTIM_TOKEN}")
fi

echo "smoke-engage-red-vs-blue: victim=${V}"

curl -fsS "${AUTH_HDR[@]}" "${V}/health" | grep -q '"ok":true' || { echo "health failed"; exit 1; }
curl -fsS "${AUTH_HDR[@]}" "${V}/api/tools" | grep -q 'nmap_scan' || { echo "tools list failed"; exit 1; }

burst_get() {
  local i
  for i in $(seq 1 40); do
    curl -fsS "${AUTH_HDR[@]}" "${V}/health" >/dev/null &
    curl -fsS "${AUTH_HDR[@]}" "${V}/api/tools" >/dev/null &
  done
  wait
}

burst_bad_tool() {
  local i
  for i in $(seq 1 60); do
    # Unknown tool: fast 404 path, no subprocess execution.
    curl -sS -o /dev/null -w "%{http_code}" "${AUTH_HDR[@]}" -X POST "${V}/api/tools/zzzz_lab_unknown_tool_${i}" \
      -H 'Content-Type: application/json' \
      -d '{}' >/dev/null &
  done
  wait
}

bad_json_posts() {
  local code
  code="$(curl -sS -o /dev/null -w "%{http_code}" "${AUTH_HDR[@]}" -X POST "${V}/api/tools/nmap_scan" \
    -H 'Content-Type: application/json' -d '{not json')"
  [[ "$code" == "4"* ]] || { echo "want 4xx on bad json, got ${code}"; exit 1; }
}

wrong_content_type() {
  local code
  code="$(curl -sS -o /dev/null -w "%{http_code}" "${AUTH_HDR[@]}" -X POST "${V}/api/tools/nmap_scan" \
    -H 'Content-Type: text/plain' -d '{"target":"127.0.0.1"}')"
  [[ "$code" == "4"* ]] || { echo "want 4xx on wrong content-type, got ${code}"; exit 1; }
}

oversized_body() {
  local code
  code="$(
    python3 -c 'import json; print(json.dumps({"target":"127.0.0.1","parameters":{"pad":"y"*(200*1024)}}))' \
      | curl -sS -o /dev/null -w "%{http_code}" "${AUTH_HDR[@]}" \
        -X POST "${V}/api/tools/nmap_scan" -H 'Content-Type: application/json' --data-binary @-
  )"
  [[ "$code" == "4"* || "$code" == "5"* ]] || { echo "oversized JSON: want 4xx/5xx, got ${code}"; exit 1; }
}

burst_get
burst_bad_tool
bad_json_posts
wrong_content_type
oversized_body

curl -fsS "${AUTH_HDR[@]}" "${V}/health" | grep -q '"ok":true' || { echo "victim unhealthy after harness"; exit 1; }

echo "OK engage red-vs-blue harness (victim still healthy)"
