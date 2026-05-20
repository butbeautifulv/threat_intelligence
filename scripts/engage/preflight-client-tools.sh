#!/usr/bin/env bash
# Preflight: verify a HexStrike-style toolset exists on PATH for client-native Engage.
# Does not install packages. See docs/engage-client-dependencies.md
# Profiles match scripts/ops/engage-tools-packages.yaml (requires PyYAML for dynamic list).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
YAML="${ENGAGE_TOOLS_PACKAGES_YAML:-${ROOT}/scripts/ops/engage-tools-packages.yaml}"
PROFILE="${ENGAGE_PREFLIGHT_PROFILE:-recommended}"
JSON_OUT=0

usage() {
  echo "Usage: $0 [--profile minimal|recommended|full] [--json]" >&2
  echo "Env: ENGAGE_PREFLIGHT_PROFILE, ENGAGE_TOOLS_PACKAGES_YAML" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      PROFILE="${2:-}"
      shift 2 || usage
      ;;
    --json) JSON_OUT=1; shift ;;
    -h|--help) usage ;;
    *) echo "unknown option: $1" >&2; usage ;;
  esac
done

MISSING=()
PRESENT=()

resolve_tools() {
  if [[ -f "$YAML" ]]; then
    PROFILE="$PROFILE" YAML="$YAML" python3 - <<'PY' 2>/dev/null || true
import os, sys, yaml
path = os.environ.get("YAML", "")
profile = os.environ.get("PROFILE", "recommended")
try:
    with open(path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)
    for name in data.get("profiles", {}).get(profile, []):
        print(name)
except Exception:
    sys.exit(1)
PY
    return
  fi
  echo ""
}

TOOLS=()
while IFS= read -r line; do
  [[ -n "$line" ]] && TOOLS+=("$line")
done < <(resolve_tools)

if ((${#TOOLS[@]} == 0)); then
  case "$PROFILE" in
    minimal) TOOLS=(nmap httpx nuclei) ;;
    full|recommended)
      TOOLS=(nmap masscan httpx nuclei subfinder amass gobuster feroxbuster ffuf sqlmap nikto)
      [[ "$PROFILE" == "full" ]] && TOOLS+=(hydra trivy)
      ;;
    *) echo "preflight-client-tools: unknown profile: $PROFILE" >&2; exit 1 ;;
  esac
fi

check() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    MISSING+=("$name")
  else
    PRESENT+=("$name")
  fi
}

for t in "${TOOLS[@]}"; do
  check "$t"
done

json_escape() {
  python3 -c 'import json,sys; print(json.dumps(sys.argv[1]))' "$1"
}

emit_json() {
  local ok=1
  ((${#MISSING[@]} == 0)) || ok=0
  local miss_json="["
  local first=1
  for m in "${MISSING[@]}"; do
    [[ $first -eq 1 ]] || miss_json+=","
    first=0
    miss_json+=$(json_escape "$m")
  done
  miss_json+="]"
  local pres_json="["
  first=1
  for p in "${PRESENT[@]}"; do
    [[ $first -eq 1 ]] || pres_json+=","
    first=0
    pres_json+=$(json_escape "$p")
  done
  pres_json+="]"
  printf '{"ok":%s,"profile":%s,"missing":%s,"present":%s}\n' \
    "$( [[ $ok -eq 1 ]] && echo true || echo false )" \
    "$(json_escape "$PROFILE")" \
    "$miss_json" \
    "$pres_json"
}

if [[ "$JSON_OUT" -eq 1 ]]; then
  emit_json
  ((${#MISSING[@]} == 0))
  exit $?
fi

if ((${#MISSING[@]} == 0)); then
  echo "preflight-client-tools: ok profile=${PROFILE} (${#TOOLS[@]} tools present)"
  exit 0
fi

echo "preflight-client-tools: profile=${PROFILE} missing on PATH: ${MISSING[*]}" >&2
echo "Install: docs/engage-install-linux.md / docs/engage-client-dependencies.md" >&2
exit 1
