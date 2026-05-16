#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export ENGAGE_CATALOG_PATH="${ENGAGE_CATALOG_PATH:-${ROOT}/engage/serve/catalog/tools.yaml}"
export AUTH_ENABLED=0
export ENGAGE_FILES_DIR="${ENGAGE_FILES_DIR:-${TMPDIR:-/tmp}/engage-mcp-smoke-files}"
export ENGAGE_AUDIT_DIR="${ENGAGE_AUDIT_DIR:-${TMPDIR:-/tmp}/engage-mcp-smoke-audit}"
mkdir -p "${ENGAGE_FILES_DIR}" "${ENGAGE_AUDIT_DIR}"
cd "${ROOT}/engage/serve"

mcp_req() {
  local payload=$1
  local len=${#payload}
  printf 'Content-Length: %d\r\n\r\n%s' "${len}" "${payload}"
}

init_payload='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}'
list_payload='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
combined=$(mcp_req "${init_payload}"; mcp_req "${list_payload}")

# head may SIGPIPE go run; do not fail the script under pipefail
resp=$(printf '%s' "${combined}" | env GOWORK="${ROOT}/engage/go.work" timeout 15 go run ./cmd/mcp 2>/dev/null | head -c 1048576 || true)
if [[ "${resp}" != *veil-engage* ]]; then
  echo "mcp initialize failed" >&2
  exit 1
fi

tool_count=$(echo "$resp" | python3 -c '
import json, sys
raw = sys.stdin.read()
pos = 0
count = 0
while pos < len(raw):
    sep = raw.find("\r\n\r\n", pos)
    if sep < 0:
        break
    header = raw[pos:sep]
    pos = sep + 4
    cl = None
    for line in header.split("\r\n"):
        if line.lower().startswith("content-length:"):
            cl = int(line.split(":", 1)[1].strip())
    if cl is None:
        break
    body = raw[pos:pos + cl]
    pos += cl
    try:
        msg = json.loads(body)
    except json.JSONDecodeError:
        continue
    if msg.get("id") == 2 and "result" in msg:
        tools = (msg.get("result") or {}).get("tools") or []
        count = len(tools)
        break
print(count)
')

if [[ -z "${tool_count}" || "${tool_count}" -lt 150 ]]; then
  echo "mcp tools/list count=${tool_count:-0} (expected >= 150)" >&2
  exit 1
fi
echo "OK engage mcp smoke (${tool_count} tools)"
