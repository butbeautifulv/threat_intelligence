#!/usr/bin/env bash
# Print tool binaries available in the engage-runner image (or local PATH).
# P9g heavy stack: ENGAGE_RUNNER_PROFILE=full or ENGAGE_RUNNER_IMAGE=engage-runner-full
set -euo pipefail
IMAGE="${ENGAGE_RUNNER_IMAGE:-engage-runner}"
PROFILE="${ENGAGE_RUNNER_PROFILE:-}"

bins=(
  nmap masscan sqlmap nikto gobuster feroxbuster
  nuclei httpx subfinder katana naabu dnsx gau waybackurls dalfox amass ffuf
  arjun dirsearch paramspider rustscan trivy
  dnsenum fierce hydra wafw00f enum4linux enum4linux-ng sslscan testssl dirb
  whatweb nbtscan binwalk jaeles x8
  engage-python-install engage-python-exec
  # P9i: remaining catalog subprocess binaries
  anew arp correlate delete detect discover display docker dotdotpwn error exiftool
  falco foremost format graphql hakrawler hashpump install intelligent jwt libc modify
  monitor msfvenom netexec objdump one optimize pacu pause prowler pwninit pwntools
  qsreplace research responder resume ropgadget ropper rpcclient scout select server
  smbmap steghide strings terminate terrascan test threat uro volatility3 vulnerability
  wfuzz xsser xxd zap
)

# P9g: 10 headless wrappers + hydra (tier-1) = 12 catalog tools on runner-full
heavy_bins=(
  burpsuite ghidra hashcat john gdb metasploit angr radare2 volatility wpscan
)

if [ "$PROFILE" = "full" ] || [ "$IMAGE" = "engage-runner-full" ]; then
  bins+=("${heavy_bins[@]}")
fi

probe() {
  local b="$1"
  if command -v docker >/dev/null 2>&1 && docker image inspect "${IMAGE}" >/dev/null 2>&1; then
    docker run --rm --entrypoint sh "${IMAGE}" -c "command -v ${b}" >/dev/null 2>&1
  else
    command -v "${b}" >/dev/null 2>&1
  fi
}

for b in "${bins[@]}"; do
  if probe "$b"; then
    echo "${b}"
  fi
done
