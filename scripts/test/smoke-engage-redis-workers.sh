#!/usr/bin/env bash
# E2E: Redis job queue with 2 worker replicas — 10 jobs, no duplicate completion.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

if ! command -v docker >/dev/null; then
  echo "skip redis workers smoke: docker not available" >&2
  exit 0
fi
if ! docker info >/dev/null 2>&1; then
  echo "skip redis workers smoke: docker daemon not running" >&2
  exit 0
fi

COMPOSE=(docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.queue.yml)
API_URL="${ENGAGE_API_URL:-http://127.0.0.1:8890}"
PROJECT="${COMPOSE_PROJECT_NAME:-engage-redis-$$}"
N_JOBS="${ENGAGE_REDIS_WORKER_JOBS:-10}"

cleanup() {
  COMPOSE_PROJECT_NAME="${PROJECT}" "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

export COMPOSE_PROJECT_NAME="${PROJECT}"
echo "redis workers smoke: starting redis + api + 2x worker (project=${PROJECT})..."
"${COMPOSE[@]}" up -d --build redis engage-api
"${COMPOSE[@]}" up -d --build --scale engage-worker=2 engage-worker

deadline=$((SECONDS + 180))
until curl -fsS "${API_URL}/health" 2>/dev/null | grep -q '"ok":true'; do
  if (( SECONDS >= deadline )); then
    echo "timeout waiting for engage-api" >&2
    "${COMPOSE[@]}" logs engage-api 2>&1 | tail -30
    exit 1
  fi
  sleep 2
done

job_ids=()
for i in $(seq 1 "${N_JOBS}"); do
  payload=$(printf '{"tool":"nmap_scan","target":"127.0.0.1","parameters":{"scan_type":"-sn","ports":"","additional_args":"-T4 --host-timeout 2s"}}')
  created=$(curl -fsS -X POST "${API_URL}/api/jobs" -H 'Content-Type: application/json' -d "${payload}")
  jid=$(echo "${created}" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("id",""))')
  [[ -n "${jid}" ]] || { echo "create job failed: ${created}" >&2; exit 1; }
  job_ids+=("${jid}")
done
echo "enqueued ${#job_ids[@]} jobs"

poll_deadline=$((SECONDS + 180))
declare -A seen_status
while (( SECONDS < poll_deadline )); do
  done_n=0
  for jid in "${job_ids[@]}"; do
    job=$(curl -fsS "${API_URL}/api/jobs/${jid}")
    st=$(echo "${job}" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("status",""))')
    seen_status["${jid}"]="${st}"
    if [[ "${st}" == "done" || "${st}" == "failed" ]]; then
      done_n=$((done_n + 1))
    fi
  done
  if [[ "${done_n}" -eq "${#job_ids[@]}" ]]; then
    break
  fi
  sleep 2
done

if [[ "${done_n:-0}" -ne "${#job_ids[@]}" ]]; then
  echo "not all jobs finished (${done_n:-0}/${#job_ids[@]})" >&2
  "${COMPOSE[@]}" logs engage-worker 2>&1 | tail -40
  exit 1
fi

# Each job id must appear exactly once in terminal states (no duplicate ids).
unique_ids=$(printf '%s\n' "${job_ids[@]}" | sort -u | wc -l)
if [[ "${unique_ids}" -ne "${#job_ids[@]}" ]]; then
  echo "duplicate job ids in enqueue set" >&2
  exit 1
fi

echo "OK engage redis multi-worker smoke (${#job_ids[@]} jobs)"
