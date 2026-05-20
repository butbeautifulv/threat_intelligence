#!/usr/bin/env bash
# Install Engage client-native CLI dependencies using the distro package manager.
# Data: scripts/ops/engage-tools-packages.yaml (requires Python PyYAML).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
YAML="${ENGAGE_TOOLS_PACKAGES_YAML:-${ROOT}/scripts/ops/engage-tools-packages.yaml}"
SOURCES_YAML="${ENGAGE_TOOLS_SOURCES_YAML:-${ROOT}/scripts/ops/engage-tools-sources.yaml}"
PROFILE="${ENGAGE_INSTALL_PROFILE:-recommended}"
INSTALL_POLICY="${ENGAGE_INSTALL_POLICY:-repo-first}"
DO_PLAN=0
DO_YES=0
DO_FALLBACK=0
MISSING_FILE=""

usage() {
  echo "Usage: $0 [--profile minimal|recommended|full] [--plan|--yes] [--policy repo-first|upstream-fallback|kali-fallback|full-auto] [--fallback] [--missing-file FILE]" >&2
  echo "  --plan   print install commands only" >&2
  echo "  --yes    run package manager (needs root/sudo on most distros)" >&2
  echo "  --fallback   install missing tools from upstream (go/cargo) when repo packages unavailable" >&2
  echo "  --policy     install policy (default: repo-first)" >&2
  echo "  --missing-file FILE   newline-delimited tool list to prioritize for fallback" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      PROFILE="${2:-}"
      shift 2 || usage
      ;;
    --plan) DO_PLAN=1; shift ;;
    --yes) DO_YES=1; shift ;;
    --fallback) DO_FALLBACK=1; shift ;;
    --policy)
      INSTALL_POLICY="${2:-}"
      shift 2 || usage
      ;;
    --missing-file)
      MISSING_FILE="${2:-}"
      shift 2 || usage
      ;;
    -h|--help) usage ;;
    *) usage ;;
  esac
done

if [[ "$DO_FALLBACK" -eq 1 && "$INSTALL_POLICY" == "repo-first" ]]; then
  INSTALL_POLICY="upstream-fallback"
fi

if [[ "$DO_PLAN" -eq 0 && "$DO_YES" -eq 0 ]]; then
  echo "Specify --plan or --yes" >&2
  usage
fi
if [[ "$DO_PLAN" -eq 1 && "$DO_YES" -eq 1 ]]; then
  echo "Use only one of --plan or --yes" >&2
  exit 1
fi

if [[ ! -f "$YAML" ]]; then
  echo "Missing YAML: $YAML" >&2
  exit 1
fi
if [[ "$DO_FALLBACK" -eq 1 && ! -f "$SOURCES_YAML" ]]; then
  echo "Missing sources YAML: $SOURCES_YAML" >&2
  exit 1
fi

case "$INSTALL_POLICY" in
  repo-first|upstream-fallback|kali-fallback|full-auto) ;;
  *)
    echo "invalid --policy ${INSTALL_POLICY}" >&2
    exit 1
    ;;
esac

detect_pm() {
  if [[ -r /etc/os-release ]]; then
    # shellcheck source=/dev/null
    . /etc/os-release
  else
    echo "Cannot read /etc/os-release" >&2
    exit 1
  fi
  case "${ID:-}" in
    debian|ubuntu|linuxmint|pop) echo apt; return ;;
    fedora|rhel|centos|rocky|almalinux) echo dnf; return ;;
    arch|manjaro) echo pacman; return ;;
    opensuse*|suse*) echo zypper; return ;;
    alpine) echo apk; return ;;
  esac
  case "${ID_LIKE:-}" in
    *debian*|*ubuntu*) echo apt; return ;;
    *fedora*|*rhel*) echo dnf; return ;;
    *arch*) echo pacman; return ;;
    *suse*) echo zypper; return ;;
  esac
  echo "Unsupported distro ID=${ID:-} ID_LIKE=${ID_LIKE:-}" >&2
  exit 1
}

PM="$(detect_pm)"

readarray -t TOOLS < <(PROFILE="$PROFILE" YAML="$YAML" python3 - <<'PY'
import os, sys, yaml
profile = os.environ["PROFILE"]
path = os.environ["YAML"]
with open(path, "r", encoding="utf-8") as f:
    data = yaml.safe_load(f)
if "profiles" not in data or profile not in data["profiles"]:
    print(f"unknown profile: {profile}", file=sys.stderr)
    sys.exit(1)
for name in data["profiles"][profile]:
    print(name)
PY
)

readarray -t PKGS < <(PROFILE="$PROFILE" PM="$PM" YAML="$YAML" python3 - <<'PY'
import os, sys, yaml
pm = os.environ["PM"]
profile = os.environ["PROFILE"]
path = os.environ["YAML"]
with open(path, "r", encoding="utf-8") as f:
    data = yaml.safe_load(f)
if "profiles" not in data or profile not in data["profiles"]:
    print(f"unknown profile: {profile}", file=sys.stderr)
    sys.exit(1)
tools = data.get("tools", {})
seen = []
for name in data["profiles"][profile]:
    meta = tools.get(name) or {}
    pkgs = meta.get(pm)
    if pkgs is None:
        print(f"tool {name}: no {pm} mapping", file=sys.stderr)
        continue
    if not pkgs:
        print(f"tool {name}: empty {pm} list (manual install)", file=sys.stderr)
        continue
    for p in pkgs:
        if p not in seen:
            seen.append(p)
            print(p)
PY
)

if ((${#PKGS[@]} == 0)); then
  echo "No packages resolved for profile=$PROFILE pm=$PM (see stderr)" >&2
  exit 1
fi

UNAVAILABLE_PKGS=()

filter_apt_available() {
  local available=()
  local p candidate
  for p in "${PKGS[@]}"; do
    candidate="$(apt-cache policy "$p" 2>/dev/null | awk '/Candidate:/ {print $2; exit}')"
    if [[ -z "${candidate:-}" || "$candidate" == "(none)" ]]; then
      UNAVAILABLE_PKGS+=("$p")
      continue
    fi
    available+=("$p")
  done
  PKGS=("${available[@]}")
}

if [[ "$PM" == "apt" ]]; then
  filter_apt_available
  if ((${#UNAVAILABLE_PKGS[@]} > 0)); then
    echo "install-engage-host-tools: apt unavailable packages (manual/alternate repo needed): ${UNAVAILABLE_PKGS[*]}" >&2
  fi
fi

if ((${#PKGS[@]} == 0)); then
  echo "No installable packages available for profile=$PROFILE pm=$PM" >&2
  exit 1
fi

run_apt() {
  if [[ "$DO_PLAN" -eq 1 ]]; then
    echo "sudo apt-get update && sudo apt-get install -y ${PKGS[*]}"
    return
  fi
  sudo apt-get update
  sudo apt-get install -y "${PKGS[@]}"
}

run_dnf() {
  if [[ "$DO_PLAN" -eq 1 ]]; then
    echo "sudo dnf install -y ${PKGS[*]}"
    return
  fi
  sudo dnf install -y "${PKGS[@]}"
}

run_pacman() {
  if [[ "$DO_PLAN" -eq 1 ]]; then
    echo "sudo pacman -S --noconfirm --needed ${PKGS[*]}"
    return
  fi
  sudo pacman -S --noconfirm --needed "${PKGS[@]}"
}

run_zypper() {
  if [[ "$DO_PLAN" -eq 1 ]]; then
    echo "sudo zypper install -y ${PKGS[*]}"
    return
  fi
  sudo zypper install -y "${PKGS[@]}"
}

run_apk() {
  if [[ "$DO_PLAN" -eq 1 ]]; then
    echo "sudo apk add ${PKGS[*]}"
    return
  fi
  sudo apk add "${PKGS[@]}"
}

case "$PM" in
  apt) run_apt ;;
  dnf) run_dnf ;;
  pacman) run_pacman ;;
  zypper) run_zypper ;;
  apk) run_apk ;;
  *) echo "internal error: pm=$PM" >&2; exit 1 ;;
esac

KALI_MISSING_PKGS=()
fallback_candidates() {
  local set=()
  local t
  if [[ -n "$MISSING_FILE" && -f "$MISSING_FILE" ]]; then
    while IFS= read -r t; do
      [[ -n "$t" ]] && set+=("$t")
    done < "$MISSING_FILE"
  else
    set=("${TOOLS[@]}")
  fi
  printf '%s\n' "${set[@]}"
}

resolve_fallback_method() {
  local tool="$1"
  TOOL="$tool" SOURCES="$SOURCES_YAML" python3 - <<'PY'
import os, yaml
tool = os.environ["TOOL"]
path = os.environ["SOURCES"]
with open(path, "r", encoding="utf-8") as f:
    data = yaml.safe_load(f) or {}
meta = (data.get("tools") or {}).get(tool) or {}
binary = meta.get("binary", tool)
methods = meta.get("preferred_install_methods") or []
method = ""
target = ""
for m in methods:
    if isinstance(m, str) and ":" in m:
        left, right = m.split(":", 1)
        if left in ("go", "cargo"):
            method = left
            target = right
            break
repo = meta.get("upstream_repo", "")
print(f"{binary}|{method}|{target}|{repo}")
PY
}

run_fallback_install() {
  local item binary method target repo
  local missing_count=0
  while IFS= read -r item; do
    [[ -n "$item" ]] || continue
    IFS="|" read -r binary method target repo < <(resolve_fallback_method "$item")
    if command -v "$binary" >/dev/null 2>&1; then
      continue
    fi
    if [[ -z "$method" || -z "$target" ]]; then
      echo "fallback: no upstream method for tool=$item (repo=${repo})" >&2
      missing_count=$((missing_count + 1))
      continue
    fi
    if [[ "$DO_PLAN" -eq 1 ]]; then
      if [[ "$method" == "go" ]]; then
        echo "go install ${target}    # ${repo}"
      else
        echo "cargo install ${target}    # ${repo}"
      fi
      continue
    fi
    if [[ "$method" == "go" ]]; then
      if ! command -v go >/dev/null 2>&1; then
        echo "fallback: go not found for ${item}" >&2
        missing_count=$((missing_count + 1))
        continue
      fi
      go install "${target}"
    else
      if ! command -v cargo >/dev/null 2>&1; then
        echo "fallback: cargo not found for ${item}" >&2
        missing_count=$((missing_count + 1))
        continue
      fi
      cargo install "${target}"
    fi
    if command -v "$binary" >/dev/null 2>&1; then
      echo "fallback: installed ${item} via ${method} (${repo})"
    else
      echo "fallback: install command ran but binary still missing: ${binary}" >&2
      missing_count=$((missing_count + 1))
    fi
  done < <(fallback_candidates)
  return $missing_count
}

collect_kali_fallback_pkgs() {
  local p
  KALI_MISSING_PKGS=()
  if [[ "$PM" != "apt" ]]; then
    return
  fi
  for p in "${UNAVAILABLE_PKGS[@]-}"; do
    KALI_MISSING_PKGS+=("$p")
  done
}

run_kali_fallback() {
  local allowlist=""
  local p
  if [[ "$PM" != "apt" ]]; then
    echo "kali-fallback: skip (non-apt distro manager: $PM)" >&2
    return 0
  fi
  collect_kali_fallback_pkgs
  if ((${#KALI_MISSING_PKGS[@]} == 0)); then
    return 0
  fi
  for p in "${KALI_MISSING_PKGS[@]}"; do
    allowlist+="${p} "
  done
  if [[ "$DO_PLAN" -eq 1 ]]; then
    echo "./scripts/ops/install-engage-kali-fallback.sh \"${allowlist% }\""
    echo "sudo apt-get install -y ${allowlist% }"
    return 0
  fi
  "${ROOT}/scripts/ops/install-engage-kali-fallback.sh" "${allowlist% }"
  sudo apt-get install -y "${KALI_MISSING_PKGS[@]}" || true
}

if [[ "$INSTALL_POLICY" == "upstream-fallback" || "$INSTALL_POLICY" == "full-auto" ]]; then
  run_fallback_install || true
fi
if [[ "$INSTALL_POLICY" == "kali-fallback" || "$INSTALL_POLICY" == "full-auto" ]]; then
  run_kali_fallback || true
fi

echo "install-engage-host-tools: done (pm=$PM profile=$PROFILE policy=$INSTALL_POLICY)"
