#!/usr/bin/env bash
# Engage hardening self-test — in-process + static audits only (safe on developer host).
# Does NOT run exploits, port scans, or attacks against the host OS.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"
cd "${VEIL_ROOT}"

log() { printf '[engage-hardening] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

log "unit: security package self-test"
(
  cd engage/serve
  env GOWORK="$(dirname "$(pwd)")/go.work" go test ./internal/security/... -count=1 -v
)

log "unit: command guard + config security"
(
  cd engage/serve
  env GOWORK="$(dirname "$(pwd)")/go.work" go test ./internal/usecase/command/... ./internal/config/... -count=1
)

log "static: compose/profile audit"
python3 "${VEIL_ROOT}/scripts/engage/hardening-compose-audit.py"

log "static: framework control catalog (JCSF/DAF/OWASP mappings)"
python3 "${VEIL_ROOT}/scripts/engage/hardening-framework-audit.py"

log "static: framework control catalog (JCSF/DAF/OWASP mappings)"
python3 "${VEIL_ROOT}/scripts/engage/hardening-framework-audit.py"

if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  log "docker: secure overlay smoke (optional)"
  chmod +x ./scripts/test/smoke-engage-secure.sh
  ENGAGE_ENV=prod ENGAGE_DENY_RAW_COMMAND=1 ./scripts/test/smoke-engage-secure.sh
else
  log "SKIP docker secure smoke (no daemon)"
fi

log "OK engage hardening self-test"
exit 0
