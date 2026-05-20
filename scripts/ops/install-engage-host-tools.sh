#!/usr/bin/env bash
# Install Engage client-native CLI dependencies using the distro package manager.
# Data: scripts/ops/engage-tools-packages.yaml (requires Python PyYAML).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
YAML="${ENGAGE_TOOLS_PACKAGES_YAML:-${ROOT}/scripts/ops/engage-tools-packages.yaml}"
PROFILE="${ENGAGE_INSTALL_PROFILE:-recommended}"
DO_PLAN=0
DO_YES=0

usage() {
  echo "Usage: $0 [--profile minimal|recommended|full] [--plan|--yes]" >&2
  echo "  --plan   print install commands only" >&2
  echo "  --yes    run package manager (needs root/sudo on most distros)" >&2
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
    -h|--help) usage ;;
    *) usage ;;
  esac
done

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

echo "install-engage-host-tools: done (pm=$PM profile=$PROFILE)"
