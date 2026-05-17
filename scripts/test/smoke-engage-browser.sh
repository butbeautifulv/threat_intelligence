#!/usr/bin/env bash
# Browser sidecar smoke: POST browser_agent_inspect via API when stack is up.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"
BROWSER_URL="${DISCOVERY_BROWSER_URL:-${ENGAGE_BROWSER_URL:-http://127.0.0.1:8920}}"

if ! curl -fsS "${BROWSER_URL}/health" 2>/dev/null | grep -q '"ok":true'; then
  echo "skip browser smoke: sidecar not reachable at ${BROWSER_URL}" >&2
  exit 0
fi
if ! curl -fsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; then
  echo "skip browser smoke: engage-api not reachable" >&2
  exit 0
fi

export DISCOVERY_BROWSER_URL="${BROWSER_URL}"
export ENGAGE_BROWSER_URL="${BROWSER_URL}"
body='{"target":"https://example.com","parameters":{"wait_time":"3"}}'
resp=$(curl -fsS -X POST "${API_URL}/api/tools/browser_agent_inspect" \
  -H 'Content-Type: application/json' \
  -d "${body}")
echo "${resp}" | grep -qE '"page_info"|"title"|"security_analysis"' || {
  echo "unexpected response: ${resp}" >&2
  exit 1
}

dash=$(curl -fsS "${API_URL}/api/processes/dashboard")
echo "${dash}" | grep -q '"system_load"' || {
  echo "dashboard missing system_load: ${dash}" >&2
  exit 1
}

echo "OK engage browser smoke"
