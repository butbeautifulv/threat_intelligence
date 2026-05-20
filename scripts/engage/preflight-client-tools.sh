#!/usr/bin/env bash
# Preflight: verify a HexStrike-style toolset exists on PATH for client-native Engage.
# Does not install packages. See docs/engage/engage-client-dependencies.md
# Profiles match scripts/ops/engage-tools-packages.yaml (requires PyYAML for dynamic list).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
YAML="${ENGAGE_TOOLS_PACKAGES_YAML:-${ROOT}/scripts/ops/engage-tools-packages.yaml}"
SOURCES_YAML="${ENGAGE_TOOLS_SOURCES_YAML:-${ROOT}/scripts/ops/engage-tools-sources.yaml}"
CORE_YAML="${ENGAGE_CORE_TOOLS_YAML:-${ROOT}/scripts/ops/engage-core-tools.yaml}"
PROFILE="${ENGAGE_PREFLIGHT_PROFILE:-recommended}"
INSTALL_POLICY="${ENGAGE_INSTALL_POLICY:-repo-first}"
JSON_OUT=0
EMIT_MISSING=0
EMIT_INSTALL_PLAN=0

usage() {
  echo "Usage: $0 [--profile minimal|recommended|full|core47] [--json] [--emit-missing] [--emit-install-plan] [--policy POLICY]" >&2
  echo "Env: ENGAGE_PREFLIGHT_PROFILE, ENGAGE_TOOLS_PACKAGES_YAML, ENGAGE_TOOLS_SOURCES_YAML, ENGAGE_CORE_TOOLS_YAML" >&2
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
    --emit-install-plan) EMIT_INSTALL_PLAN=1; shift ;;
    --policy)
      INSTALL_POLICY="${2:-}"
      shift 2 || usage
      ;;
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
    core47)
      TOOLS=(nmap masscan rustscan amass subfinder nuclei fierce dnsenum autorecon theharvester responder netexec enum4linux-ng gobuster feroxbuster dirsearch ffuf dirb httpx katana nikto sqlmap wpscan arjun paramspider dalfox wafw00f hydra john hashcat medusa patator crackmapexec evil-winrm hash-identifier ophcrack gdb radare2 binwalk ghidra checksec strings objdump volatility3 foremost steghide exiftool)
      ;;
    *) echo "preflight-client-tools: unknown profile: $PROFILE" >&2; exit 1 ;;
  esac
fi

resolve_binary() {
  local name="$1"
  TOOL="$name" CORE_YAML="$CORE_YAML" SOURCES_YAML="$SOURCES_YAML" python3 - <<'PY'
import os, pathlib, yaml
name = os.environ["TOOL"]
for key in ("CORE_YAML", "SOURCES_YAML"):
    p = pathlib.Path(os.environ.get(key, ""))
    if not p.is_file():
        continue
    with p.open("r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    tools = data.get("tools") or {}
    meta = tools.get(name)
    if isinstance(meta, dict):
        b = (meta.get("binary") or "").strip()
        if b:
            print(b)
            raise SystemExit(0)
print(name)
PY
}

check() {
  local name="$1"
  local bin
  bin="$(resolve_binary "$name")"
  if command -v "$bin" >/dev/null 2>&1; then
    PRESENT+=("$name")
    return
  fi
  # fallback to direct name for legacy compatibility
  if command -v "$name" >/dev/null 2>&1; then
    PRESENT+=("$name")
  else
    MISSING+=("$name")
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

if [[ "$EMIT_INSTALL_PLAN" -eq 1 ]]; then
  echo "./scripts/engage/preflight-client-tools.sh --profile ${PROFILE} --emit-missing | ./scripts/ops/install-engage-host-tools.sh --plan --profile ${PROFILE} --policy ${INSTALL_POLICY} --missing-file /dev/stdin"
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
echo "Install: docs/engage/engage-install-linux.md / docs/engage/engage-client-dependencies.md" >&2
exit 1
