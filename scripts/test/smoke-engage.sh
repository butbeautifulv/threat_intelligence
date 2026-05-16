#!/usr/bin/env bash
set -euo pipefail
API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"
curl -fsS "${API_URL}/health" | grep -q '"ok":true' || { echo "health failed"; exit 1; }
curl -fsS "${API_URL}/api/tools" | grep -q 'nmap_scan' || { echo "tools list failed"; exit 1; }
echo "OK engage smoke"
