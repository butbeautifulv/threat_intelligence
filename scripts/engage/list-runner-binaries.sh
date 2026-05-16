#!/usr/bin/env bash
# Print tool binaries available in the engage-runner image (or local PATH).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
IMAGE="${ENGAGE_RUNNER_IMAGE:-engage-runner}"

bins=(
  nmap masscan sqlmap nikto gobuster feroxbuster
  nuclei httpx subfinder katana naabu dnsx gau waybackurls dalfox amass ffuf
  arjun dirsearch paramspider rustscan trivy
  dnsenum fierce hydra wafw00f enum4linux enum4linux-ng sslscan testssl dirb
  whatweb nbtscan binwalk jaeles x8
)

if command -v docker >/dev/null 2>&1 && docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  for b in "${bins[@]}"; do
    if docker run --rm --entrypoint sh "${IMAGE}" -c "command -v ${b}" >/dev/null 2>&1; then
      echo "${b}"
    fi
  done
else
  for b in "${bins[@]}"; do
    if command -v "${b}" >/dev/null 2>&1; then
      echo "${b}"
    fi
  done
fi
