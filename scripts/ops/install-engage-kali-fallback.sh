#!/usr/bin/env bash
# Configure Kali apt fallback with pinning for selected packages only.
set -euo pipefail

LIST_FILE="/etc/apt/sources.list.d/engage-kali-fallback.list"
PREF_FILE="/etc/apt/preferences.d/engage-kali-fallback.pref"
DIST="${ENGAGE_KALI_DIST:-kali-rolling}"
MIRROR="${ENGAGE_KALI_MIRROR:-http://http.kali.org/kali}"
ALLOWLIST="${1:-}"

if [[ -z "${ALLOWLIST}" ]]; then
  echo "usage: $0 'pkg1 pkg2 ...'" >&2
  exit 1
fi

sudo install -d -m 0755 "$(dirname "$LIST_FILE")" "$(dirname "$PREF_FILE")"
echo "deb ${MIRROR} ${DIST} main contrib non-free non-free-firmware" | sudo tee "$LIST_FILE" >/dev/null

{
  echo "Package: *"
  echo "Pin: release n=${DIST}"
  echo "Pin-Priority: 50"
  echo
  for p in ${ALLOWLIST}; do
    echo "Package: ${p}"
    echo "Pin: release n=${DIST}"
    echo "Pin-Priority: 700"
    echo
  done
} | sudo tee "$PREF_FILE" >/dev/null

sudo apt-get update
echo "configured Kali fallback repo with pinned allowlist: ${ALLOWLIST}"
