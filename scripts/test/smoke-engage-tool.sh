#!/usr/bin/env bash
# Opt-in smoke: POST a live catalog tool (nmap_scan on 127.0.0.1, -sn).
# Skip in CI without runner: ENGAGE_SKIP_TOOL_SMOKE=1
set -euo pipefail

if [[ "${ENGAGE_SKIP_TOOL_SMOKE:-}" == "1" ]]; then
  echo "SKIP engage tool smoke (ENGAGE_SKIP_TOOL_SMOKE=1)"
  exit 0
fi

API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"
TOOL="${ENGAGE_SMOKE_TOOL:-nmap_scan}"
TARGET="${ENGAGE_SMOKE_TARGET:-127.0.0.1}"
TIMEOUT="${ENGAGE_SMOKE_TIMEOUT:-120}"

body=$(cat <<EOF
{"target":"${TARGET}","parameters":{"scan_type":"-sn","ports":"","additional_args":"-T4"}}
EOF
)

resp=$(curl -fsS --max-time "${TIMEOUT}" -X POST "${API_URL}/api/tools/${TOOL}" \
  -H 'Content-Type: application/json' \
  -d "${body}") || {
  echo "tool POST failed (is engage-api up? try: make test-engage-smoke)"
  exit 1
}

if echo "${resp}" | grep -q '"success":true'; then
  echo "OK engage tool smoke (${TOOL} on ${TARGET})"
  exit 0
fi

# Binary missing or runner misconfigured — warn but do not fail minimal CI
if echo "${resp}" | grep -qiE 'not found|executable|no such file|docker|connection refused'; then
  echo "SKIP engage tool smoke: ${resp}" >&2
  exit 0
fi

echo "engage tool smoke failed: ${resp}" >&2
exit 1
