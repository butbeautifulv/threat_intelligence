#!/usr/bin/env bash
# Best-effort CI matrix: smoke up to 10 catalog tools when binaries exist on PATH.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SMOKE="${ROOT}/scripts/test/smoke-engage-tool.sh"
chmod +x "${SMOKE}"

tools=(
  nmap_scan:127.0.0.1
  nuclei_scan:https://example.com
  httpx_probe:https://example.com
  subfinder_scan:example.com
  gobuster_scan:http://127.0.0.1
  nikto_scan:127.0.0.1
  ffuf_scan:http://127.0.0.1
  arjun_scan:https://example.com
  rustscan_fast_scan:127.0.0.1
  trivy_scan:alpine:latest
  sqlmap_scan:http://127.0.0.1
  api_fuzzer:https://example.com
  graphql_scanner:https://example.com
)

ran=0
for entry in "${tools[@]}"; do
  tool="${entry%%:*}"
  target="${entry#*:}"
  bin="${tool%%_*}"
  if ! command -v "${bin}" >/dev/null 2>&1; then
    echo "skip ${tool}: ${bin} not on PATH" >&2
    continue
  fi
  echo "smoke ${tool} -> ${target}"
  "${SMOKE}" "${tool}" "${target}" || true
  ran=$((ran + 1))
done
if [[ "${ran}" -eq 0 ]]; then
  echo "skip tool matrix: no supported binaries on PATH" >&2
  exit 0
fi
echo "OK engage tool matrix (${ran} tools exercised)"
