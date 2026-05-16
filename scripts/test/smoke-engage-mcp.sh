#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export ENGAGE_CATALOG_PATH="${ENGAGE_CATALOG_PATH:-${ROOT}/engage/serve/catalog/tools.yaml}"
export AUTH_ENABLED=0
cd "${ROOT}/engage/serve"
payload='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}'
len=${#payload}
req="Content-Length: ${len}\r\n\r\n${payload}"
resp=$(printf '%b' "$req" | env GOWORK="${ROOT}/engage/go.work" timeout 5 go run ./cmd/mcp 2>/dev/null | head -c 4096) || resp=""
echo "$resp" | grep -q veil-engage || { echo "mcp initialize failed"; exit 1; }
echo "OK engage mcp smoke"
