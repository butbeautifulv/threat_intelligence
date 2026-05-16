#!/usr/bin/env bash
# E2E: api + worker + runner (docker exec) — async job via POST /api/jobs.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null; then
  echo "skip compose smoke: docker not available" >&2
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "skip compose smoke: docker daemon not running" >&2
  exit 0
fi

COMPOSE=(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.runner.yml --profile runner)
API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"
PROJECT="${COMPOSE_PROJECT_NAME:-engage-smoke-$$}"

cleanup() {
  COMPOSE_PROJECT_NAME="${PROJECT}" "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

export COMPOSE_PROJECT_NAME="${PROJECT}"
echo "compose smoke: building and starting api worker runner (project=${PROJECT})..."
"${COMPOSE[@]}" up -d --build engage-runner engage-api engage-worker

deadline=$((SECONDS + 180))
until curl -fsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; do
  if (( SECONDS >= deadline )); then
    echo "timeout waiting for engage-api health" >&2
    "${COMPOSE[@]}" logs engage-api 2>&1 | tail -40
    exit 1
  fi
  sleep 2
done

job_payload='{"tool":"nmap_scan","target":"127.0.0.1","parameters":{"scan_type":"-sn","ports":"","additional_args":"-T4"}}'
created=$(curl -fsS -X POST "${API_URL}/api/jobs" -H 'Content-Type: application/json' -d "${job_payload}")
job_id=$(echo "${created}" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("id",""))')
if [[ -z "${job_id}" ]]; then
  echo "failed to create job: ${created}" >&2
  exit 1
fi
echo "job created: ${job_id}"

poll_deadline=$((SECONDS + 120))
status=""
while (( SECONDS < poll_deadline )); do
  job=$(curl -fsS "${API_URL}/api/jobs/${job_id}")
  status=$(echo "${job}" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("status",""))')
  if [[ "${status}" == "done" || "${status}" == "failed" ]]; then
    echo "job finished: status=${status}"
    break
  fi
  sleep 2
done

if [[ "${status}" != "done" && "${status}" != "failed" ]]; then
  echo "job did not complete in time: last status=${status}" >&2
  "${COMPOSE[@]}" logs engage-worker 2>&1 | tail -30
  exit 1
fi

chmod +x "${ROOT}/scripts/test/smoke-engage-runner-profile.sh"
"${ROOT}/scripts/test/smoke-engage-runner-profile.sh" || true

RUNNER_CTN=$(docker ps --filter "name=engage-runner" --format '{{.Names}}' | head -1 || true)
chmod +x "${ROOT}/scripts/test/smoke-engage-tool-matrix.sh"
ENGAGE_RUNNER_CONTAINER="${RUNNER_CTN}" \
  ENGAGE_TOOL_MATRIX_STRICT=1 ENGAGE_TOOL_MATRIX_MIN=30 \
  "${ROOT}/scripts/test/smoke-engage-tool-matrix.sh"

echo "OK engage compose smoke"

if [[ "${ENGAGE_COMPOSE_BENCHMARK_SKIP:-}" == "1" ]]; then
  echo "compose smoke: skipping benchmark hook (ENGAGE_COMPOSE_BENCHMARK_SKIP=1)" >&2
  exit 0
fi

chmod +x "${ROOT}/scripts/benchmark/engage-hexstrike-parity.sh"
ENGAGE_API_URL="${API_URL}" ENGAGE_BENCHMARK_EXECUTE="${ENGAGE_BENCHMARK_EXECUTE:-0}" \
  "${ROOT}/scripts/benchmark/engage-hexstrike-parity.sh" || true
echo "compose smoke: benchmark hook done (SKIP ok if api unreachable)"
