#!/usr/bin/env bash
# Deprecated alias — use smoke-discovery-e2e.sh
echo 'DEPRECATED: use ./scripts/test/smoke-discovery-e2e.sh' >&2
exec "$(cd "$(dirname "$0")" && pwd)/smoke-discovery-e2e.sh" "$@"
