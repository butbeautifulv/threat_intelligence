#!/usr/bin/env bash
# P11a: run executable-matrix inside engage-runner-full (real PATH, not host stub layer).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null 2>&1; then
  echo "skip engage executable matrix runner: docker not available" >&2
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "skip engage executable matrix runner: docker daemon not running" >&2
  exit 0
fi

IMAGE="${ENGAGE_RUNNER_FULL_MATRIX_IMAGE:-engage-runner-full-matrix}"
DOCKERFILE="${ROOT}/deploy/engage/docker/runner.Dockerfile"
GO_VERSION="${ENGAGE_MATRIX_GO_VERSION:-1.25.0}"
EXPECTED="${ENGAGE_MATRIX_EXPECTED:-158}"

echo "engage-runner-full matrix: build ${IMAGE} (target engage-runner-full)"
docker build -f "${DOCKERFILE}" --target engage-runner-full -t "${IMAGE}" "${ROOT}"

echo "engage-runner-full matrix: go run ./cmd/executable-matrix (ENGAGE_MATRIX_IN_RUNNER=1)"
out="$(mktemp)"
trap 'rm -f "${out}"' EXIT

docker run --rm --user root \
  -v "${ROOT}:/work:ro" \
  -e ENGAGE_MATRIX_IN_RUNNER=1 \
  -e GOWORK=/work/engage/go.work \
  -e GOPATH=/tmp/gopath \
  -e GOCACHE=/tmp/gocache \
  "${IMAGE}" \
  bash -ec "
    set -euo pipefail
    export PATH=/usr/local/bin:/usr/bin:/bin
    if ! command -v go >/dev/null 2>&1 || ! go version 2>/dev/null | grep -q 'go${GO_VERSION}'; then
      apt-get update -qq
      apt-get install -y -qq ca-certificates curl git >/dev/null
      curl -fsSL -o /tmp/go.tgz \"https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz\"
      rm -rf /usr/local/go
      tar -C /usr/local -xzf /tmp/go.tgz
      rm -f /tmp/go.tgz
    fi
    export PATH=/usr/local/go/bin:\${PATH}
    cd /work/engage/serve
    go run ./cmd/executable-matrix/ --root /work
  " >"${out}" 2>&1 || {
  cat "${out}" >&2
  exit 1
}

total=0
passed=0
while IFS= read -r line; do
  case "${line}" in
    MATRIX_TOTAL=*) total="${line#MATRIX_TOTAL=}" ;;
    MATRIX_PASS=*) passed="${line#MATRIX_PASS=}" ;;
  esac
done < <(grep -E '^MATRIX_(TOTAL|PASS)=' "${out}" || true)

echo "${out}" | grep -E '^MATRIX_(TOTAL|PASS)=' >&2 || true
echo "engage-runner-full executable matrix: ${passed:-?}/${total:-?} PASS"

if [ "${passed:-0}" != "${EXPECTED}" ] || [ "${total:-0}" != "${EXPECTED}" ]; then
  echo "engage-runner-full matrix: expected ${EXPECTED}/${EXPECTED}" >&2
  echo "${out}" | tail -n 40 >&2
  exit 1
fi

echo "OK engage-runner-full executable matrix (${EXPECTED}/${EXPECTED})"
