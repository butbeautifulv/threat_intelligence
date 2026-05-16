#!/usr/bin/env bash
# Smoke: Prometheus /metrics when ENGAGE_METRICS_ENABLED=1.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"

if ! curl -fsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; then
  echo "skip metrics smoke: engage-api not reachable at ${API_URL}" >&2
  exit 0
fi

body=$(curl -fsS "${API_URL}/metrics" 2>/dev/null || true)
if [[ -z "${body}" ]]; then
  echo "skip metrics smoke: /metrics not enabled (set ENGAGE_METRICS_ENABLED=1)" >&2
  exit 0
fi
echo "${body}" | grep -q 'engage_tool_runs_total' || grep -q 'engage_audit_events_total' <<<"${body}"
echo "OK engage metrics smoke"
