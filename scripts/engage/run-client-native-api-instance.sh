#!/usr/bin/env bash
# Run engage-api on the host for red-vs-blue lab: isolated ports and state dirs.
# Usage: ./scripts/engage/run-client-native-api-instance.sh victim|attacker
# See docs/engage/engage-red-blue-lab.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ROLE="${1:-${ENGAGE_LAB_ROLE:-}}"

lab_base() {
  echo "${ENGAGE_LAB_ROOT:-/tmp/engage-lab}/${1}"
}

case "$ROLE" in
  victim)
    export ENGAGE_API_LISTEN="${ENGAGE_API_LISTEN:-:8891}"
    BASE="$(lab_base victim)"
    ;;
  attacker)
    export ENGAGE_API_LISTEN="${ENGAGE_API_LISTEN:-:8890}"
    BASE="$(lab_base attacker)"
    ;;
  *)
    echo "usage: $0 victim|attacker" >&2
    echo "Env overrides: ENGAGE_LAB_ROLE, ENGAGE_LAB_ROOT, ENGAGE_API_LISTEN" >&2
    exit 1
    ;;
esac

mkdir -p "${BASE}/work" "${BASE}/files" "${BASE}/audit" "${BASE}/jobs"
export ENGAGE_RUNNER_WORKDIR="${ENGAGE_RUNNER_WORKDIR:-${BASE}/work}"
export ENGAGE_FILES_DIR="${ENGAGE_FILES_DIR:-${BASE}/files}"
export ENGAGE_AUDIT_DIR="${ENGAGE_AUDIT_DIR:-${BASE}/audit}"
export ENGAGE_JOBS_DIR="${ENGAGE_JOBS_DIR:-${BASE}/jobs}"

exec "${ROOT}/scripts/engage/run-client-native-api.sh"
