#!/usr/bin/env bash
# P9j: build engage-runner-full and verify heavy-stack binaries + minimal invocations.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null 2>&1; then
  echo "skip engage-runner-full smoke: docker not available" >&2
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "skip engage-runner-full smoke: docker daemon not running" >&2
  exit 0
fi

IMAGE="${ENGAGE_RUNNER_FULL_SMOKE_IMAGE:-engage-runner-full-smoke}"
DOCKERFILE="${ROOT}/deploy/engage/docker/runner.Dockerfile"
TOOLS=(nmap burpsuite ghidra hashcat hydra metasploit angr radare2 volatility wpscan)

echo "engage-runner-full smoke: build ${IMAGE} (target engage-runner-full)"
docker build -f "${DOCKERFILE}" --target engage-runner-full -t "${IMAGE}" "${ROOT}"

run() {
  docker run --rm --user 10001 "${IMAGE}" "$@"
}

echo "engage-runner-full smoke: which heavy-stack binaries"
missing=()
for t in "${TOOLS[@]}"; do
  if ! run sh -c "command -v ${t}" >/dev/null 2>&1; then
    missing+=("${t}")
  fi
done
if [ "${#missing[@]}" -gt 0 ]; then
  echo "engage-runner-full smoke: missing on PATH: ${missing[*]}" >&2
  exit 1
fi
for t in "${TOOLS[@]}"; do
  echo "  ok which ${t}"
done

echo "engage-runner-full smoke: minimal invocations"
run nmap --version | head -n1
echo "  ok nmap --version"
run hashcat --version | head -n1
echo "  ok hashcat --version"
run radare2 -v | head -n1
echo "  ok radare2 -v"
run metasploit 2>&1 | head -n1
echo "  ok metasploit (version stub)"
run angr /bin/true
echo "  ok angr /bin/true"
run wpscan --version 2>&1 | head -n1
echo "  ok wpscan --version"

echo "OK engage-runner-full smoke"
