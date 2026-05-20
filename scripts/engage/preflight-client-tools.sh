#!/usr/bin/env bash
# Preflight: verify a minimal HexStrike-style toolset exists on PATH for client-native Engage.
# Does not install packages. See docs/engage-client-dependencies.md
set -euo pipefail

MISSING=()
check() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    MISSING+=("$name")
  fi
}

# Core subset (network + web recon); extend as needed.
TOOLS=(nmap masscan httpx nuclei subfinder amass gobuster feroxbuster ffuf sqlmap nikto)
for t in "${TOOLS[@]}"; do
  check "$t"
done

if ((${#MISSING[@]} == 0)); then
  echo "preflight-client-tools: ok (${#TOOLS[@]} tools present)"
  exit 0
fi

echo "preflight-client-tools: missing on PATH: ${MISSING[*]}" >&2
echo "Install examples: docs/engage-client-dependencies.md" >&2
exit 1
