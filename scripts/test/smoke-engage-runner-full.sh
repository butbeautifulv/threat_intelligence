#!/usr/bin/env bash
# P9j: build engage-runner-full and verify heavy-stack binaries + minimal invocations.
# P10d: cloud security tools (prowler, scout-suite, pacu, terrascan, netexec, stubs).
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
CLOUD_BINARIES=(prowler scout pacu terrascan netexec docker)
CLOUD_STUBS=(kube-hunter kube-bench checkov clair falco kube)

echo "engage-runner-full smoke: build ${IMAGE} (target engage-runner-full)"
docker build -f "${DOCKERFILE}" --target engage-runner-full -t "${IMAGE}" "${ROOT}"

run() {
  docker run --rm --user 10001 "${IMAGE}" "$@"
}

assert_on_path() {
  local t
  for t in "$@"; do
    if ! run sh -c "command -v ${t}" >/dev/null 2>&1; then
      echo "engage-runner-full smoke: missing on PATH: ${t}" >&2
      exit 1
    fi
    echo "  ok which ${t}"
  done
}

assert_stub() {
  local bin=$1
  local out
  out="$(run "${bin}" 2>&1 | head -n1)"
  if ! echo "${out}" | grep -q '"stub":true'; then
    echo "engage-runner-full smoke: expected stub JSON from ${bin}, got: ${out}" >&2
    exit 1
  fi
  echo "  ok ${bin} (stub)"
}

try_version_or_help() {
  local bin=$1
  if run "${bin}" --version 2>&1 | head -n1 | grep -qv '"stub":true'; then
    echo "  ok ${bin} --version"
    return 0
  fi
  if run "${bin}" --help 2>&1 | head -n1 | grep -qv '"stub":true'; then
    echo "  ok ${bin} --help"
    return 0
  fi
  return 1
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

echo "engage-runner-full smoke: cloud security binaries on PATH"
assert_on_path "${CLOUD_BINARIES[@]}"

echo "engage-runner-full smoke: cloud tool invocations"
try_version_or_help prowler || assert_stub prowler
if run scout-suite --help 2>&1 | head -n1 | grep -qiE 'usage|scout|Scout'; then
  echo "  ok scout-suite --help"
else
  try_version_or_help scout || assert_stub scout
fi
if run pacu --help 2>&1 | head -n1 | grep -qiE 'usage|pacu|help'; then
  echo "  ok pacu --help"
elif run pacu 2>&1 | head -n1 | grep -qv '"stub":true'; then
  echo "  ok pacu"
else
  assert_stub pacu
fi
if run terrascan version 2>&1 | head -n1 | grep -qiE 'version|terrascan'; then
  echo "  ok terrascan version"
else
  try_version_or_help terrascan || assert_stub terrascan
fi
if run nxc --help 2>&1 | head -n1 | grep -qiE 'usage|netexec|nxc'; then
  echo "  ok nxc --help"
else
  try_version_or_help netexec || assert_stub netexec
fi
if ! run sh -c 'test -x /opt/docker-bench/docker-bench-security.sh'; then
  echo "engage-runner-full smoke: docker-bench-security script missing" >&2
  exit 1
fi
echo "  ok docker-bench-security script"
if run docker-bench-security --help 2>&1 | head -n1 | grep -qiE 'docker|bench|usage|help'; then
  echo "  ok docker-bench-security --help"
else
  echo "  ok docker-bench-security (wrapper)"
fi

echo "engage-runner-full smoke: cloud catalog stubs"
for s in "${CLOUD_STUBS[@]}"; do
  assert_stub "${s}"
done

echo "OK engage-runner-full smoke"
