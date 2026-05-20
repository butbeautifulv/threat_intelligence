#!/usr/bin/env bash
# Preflight: verify a HexStrike-style toolset exists on PATH for client-native Engage.
# Does not install packages. See docs/engage-client-dependencies.md
# Profiles match scripts/ops/engage-tools-packages.yaml (requires PyYAML for dynamic list).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
YAML="${ENGAGE_TOOLS_PACKAGES_YAML:-${ROOT}/scripts/ops/engage-tools-packages.yaml}"
SOURCES_YAML="${ENGAGE_TOOLS_SOURCES_YAML:-${ROOT}/scripts/ops/engage-tools-sources.yaml}"
PROFILE="${ENGAGE_PREFLIGHT_PROFILE:-recommended}"
JSON_OUT=0
EMIT_MISSING=0

usage() {
  echo "Usage: $0 [--profile minimal|recommended|full] [--json] [--emit-missing]" >&2
  echo "Env: ENGAGE_PREFLIGHT_PROFILE, ENGAGE_TOOLS_PACKAGES_YAML, ENGAGE_TOOLS_SOURCES_YAML" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      PROFILE="${2:-}"
      shift 2 || usage
      ;;
    --json) JSON_OUT=1; shift ;;
    --emit-missing) EMIT_MISSING=1; shift ;;
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
  local joined_missing joined_present
  joined_missing="$(printf '%s\n' "${MISSING[@]-}")"
  joined_present="$(printf '%s\n' "${PRESENT[@]-}")"
  PROFILE="$PROFILE" SOURCES_YAML="$SOURCES_YAML" MISSING_NL="$joined_missing" PRESENT_NL="$joined_present" python3 - <<'PY'
import json, os, pathlib, yaml
profile = os.environ.get("PROFILE", "recommended")
missing = [x for x in os.environ.get("MISSING_NL", "").splitlines() if x]
present = [x for x in os.environ.get("PRESENT_NL", "").splitlines() if x]
hints = {}
src_path = pathlib.Path(os.environ.get("SOURCES_YAML", ""))
if src_path.is_file():
    with src_path.open("r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    tools = data.get("tools") or {}
    for t in missing:
        meta = tools.get(t) or {}
        hints[t] = {
            "kali_tool_page": meta.get("kali_tool_page", ""),
            "kali_pkg_tracker": meta.get("kali_pkg_tracker", ""),
            "upstream_repo": meta.get("upstream_repo", ""),
            "preferred_install_methods": meta.get("preferred_install_methods", []),
        }
print(json.dumps({
    "ok": len(missing) == 0,
    "profile": profile,
    "missing": missing,
    "present": present,
    "hints": hints,
}))
PY
}

if [[ "$EMIT_MISSING" -eq 1 ]]; then
  printf '%s\n' "${MISSING[@]}"
  ((${#MISSING[@]} == 0))
  exit $?
fi

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
