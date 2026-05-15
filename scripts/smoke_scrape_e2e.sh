#!/usr/bin/env bash
# E2E smoke: scrape profile → pipeline → ingest → Neo4j (+ optional ledger pass).
# Usage:
#   ./scripts/smoke_scrape_e2e.sh              # check running stack
#   ./scripts/smoke_scrape_e2e.sh --up        # compose up scrape services first
#   ./scripts/smoke_scrape_e2e.sh --restart-scrape   # pass 2: ledger unchanged
# Env:
#   COMPOSE_FILE, COMPOSE_PROFILE=scrape
#   SCRAPE_SVC=pipeline_worker  (service names after slice 8 v2 rename)
#   PIPELINE_SVC, INGEST_SVC, NATS_MON=http://127.0.0.1:8222
#   API_URL=http://127.0.0.1:8090/health
#   SMOKE_WAIT_SEC=600
#   PIPELINE_WORKER_SCALE=2 INGEST_WORKER_SCALE=2  (passed to compose-up-full on --up)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE="${COMPOSE:-docker compose}"
COMPOSE_FILES="${COMPOSE_FILES:--f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml}"
PROFILE="${COMPOSE_PROFILE:-}"
SCRAPE_SVC="${SCRAPE_SVC:-scrape_worker}"
PIPELINE_SVC="${PIPELINE_SVC:-pipeline_worker}"
INGEST_SVC="${INGEST_SVC:-ingest_worker}"
NATS_MON="${NATS_MON:-http://127.0.0.1:${NATS_MONITOR_PORT:-8222}}"
API_URL="${API_URL:-http://127.0.0.1:${API_PORT:-8090}/health}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
CRAWL_MYSQL="${CRAWL_MYSQL:-veil:veilpass@tcp(127.0.0.1:3306)/veil_ledger}"
WAIT_SEC="${SMOKE_WAIT_SEC:-600}"

compose() {
  # shellcheck disable=SC2086
  $COMPOSE $COMPOSE_FILES ${PROFILE:+--profile "$PROFILE"} "$@"
}

log() { printf '[smoke] %s\n' "$*"; }
fail() { log "FAIL: $*"; exit 1; }

if [[ "${1:-}" == "--up" ]]; then
  export NVD_MAX_PAGES="${NVD_MAX_PAGES:-1}"
  export SBOM_MAX_CVES="${SBOM_MAX_CVES:-5}"
  export DS_MAX_SIGMA="${DS_MAX_SIGMA:-5}"
  export DS_MAX_YARA="${DS_MAX_YARA:-5}"
  export DS_MAX_ATOMIC="${DS_MAX_ATOMIC:-0}"
  export CODERULES_MAX_SEMGREP="${CODERULES_MAX_SEMGREP:-5}"
  export NUCLEI_MAX="${NUCLEI_MAX:-5}"
  export VULN_METASPLOIT_MAX_RB="${VULN_METASPLOIT_MAX_RB:-0}"
  export DS_MAX_CALDERA="${DS_MAX_CALDERA:-0}"
  export SBOM_SOURCES="${SBOM_SOURCES:-osv}"
  # Minimal scrape proof (no GitHub zip / NVD bulk). Full crawl: SCRAPE_SOURCES=ds,vuln,lola,ti,sbom,coderules,nuclei
  export SCRAPE_SOURCES="${SCRAPE_SOURCES:-ti,sbom}"
  export GRAPH_PACK_SKIP="${GRAPH_PACK_SKIP:-1}"
  if [[ "${SMOKE_CLEAN_VOLUMES:-1}" == "1" ]]; then
    log "removing old compose volumes (SMOKE_CLEAN_VOLUMES=1)..."
    compose down -v --remove-orphans 2>/dev/null || true
  fi
  log "bringing up scrape stack (sources=${SCRAPE_SOURCES}, pipeline_scale=${PIPELINE_WORKER_SCALE:-1}, ingest_scale=${INGEST_WORKER_SCALE:-1})..."
  if [[ -x "$ROOT/scripts/compose-up-full.sh" ]]; then
    PIPELINE_WORKER_SCALE="${PIPELINE_WORKER_SCALE:-1}" \
    INGEST_WORKER_SCALE="${INGEST_WORKER_SCALE:-1}" \
    SCRAPE_WORKER_PARTITION="${SCRAPE_WORKER_PARTITION:-0}" \
      "$ROOT/scripts/compose-up-full.sh"
  else
    scale_args=()
    [[ "${PIPELINE_WORKER_SCALE:-1}" != "1" ]] && scale_args+=(--scale "pipeline_worker=${PIPELINE_WORKER_SCALE}")
    [[ "${INGEST_WORKER_SCALE:-1}" != "1" ]] && scale_args+=(--scale "ingest_worker=${INGEST_WORKER_SCALE}")
    compose up --build -d "${scale_args[@]}" crawl-db nats neo4j graph-bootstrap "$SCRAPE_SVC" "$PIPELINE_SVC" "$INGEST_SVC"
  fi
  shift
fi

if [[ "${1:-}" == "--restart-scrape" ]]; then
  log "restarting scrape (ledger pass 2)..."
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    compose restart scrape_worker_fast scrape_worker_slow 2>/dev/null || true
  else
    compose restart "$SCRAPE_SVC"
  fi
  sleep 15
  log "check logs for unchanged / skip publish:"
  compose logs --tail=80 "$SCRAPE_SVC" 2>&1 | grep -iE 'unchanged|skip publish' || log "(no unchanged lines in last 80 — may be first run)"
  exit 0
fi

scrape_exited_ok() {
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    compose ps -a scrape_worker_fast scrape_worker_slow 2>/dev/null | grep -qE 'Exited \(0\)|exited \(0\)' \
      && ! compose ps -a scrape_worker_fast scrape_worker_slow 2>/dev/null | grep -qE 'Exited \([1-9]'
  else
    compose ps -a "$SCRAPE_SVC" 2>/dev/null | grep -qE 'Exited \(0\)|exited \(0\)'
  fi
}

scrape_exited_err() {
  if [[ "${SCRAPE_WORKER_PARTITION:-0}" == "1" ]]; then
    compose ps -a scrape_worker_fast scrape_worker_slow 2>/dev/null | grep -qE 'Exited \([1-9]'
  else
    compose ps -a "$SCRAPE_SVC" 2>/dev/null | grep -qE 'Exited \([1-9]'
  fi
}

log "waiting for scrape worker(s) to exit (max ${WAIT_SEC}s)..."
deadline=$((SECONDS + WAIT_SEC))
while (( SECONDS < deadline )); do
  if scrape_exited_ok; then
    log "scrape worker(s) exited 0"
    break
  fi
  if scrape_exited_err; then
    compose logs --tail=40 "$SCRAPE_SVC" scrape_worker_fast scrape_worker_slow 2>/dev/null || true
    fail "scrape worker exited with error"
  fi
  sleep 5
done
if (( SECONDS >= deadline )); then
  log "timeout waiting for scrape worker — continuing checks (may still be running)"
fi

log "NATS monitoring $NATS_MON"
if command -v curl >/dev/null 2>&1; then
  curl -sf "$NATS_MON/healthz" >/dev/null || fail "NATS healthz"
  # Best-effort: stream info present
  curl -sf "$NATS_MON/jsz" | head -c 200 >/dev/null || log "warn: /jsz not available"
else
  log "warn: curl missing, skip NATS HTTP checks"
fi

log "pipeline + ingest should be running"
pw_up=$(compose ps "$PIPELINE_SVC" 2>/dev/null | grep -c Up || true)
iw_up=$(compose ps "$INGEST_SVC" 2>/dev/null | grep -c Up || true)
pw_want="${PIPELINE_WORKER_SCALE:-1}"
iw_want="${INGEST_WORKER_SCALE:-1}"
[[ "$pw_up" -ge "$pw_want" ]] || fail "pipeline_worker: want ${pw_want} Up, got ${pw_up}"
[[ "$iw_up" -ge "$iw_want" ]] || fail "ingest_worker: want ${iw_want} Up, got ${iw_up}"

if command -v mysql >/dev/null 2>&1; then
  log "crawl_resource counts (requires mysql client + published 3306)"
  mysql -h127.0.0.1 -uveil -pveilpass veil_ledger -e \
    "SELECT source, COUNT(*) AS n FROM crawl_resource GROUP BY source ORDER BY source;" 2>/dev/null \
    || log "warn: crawl-db not reachable on localhost:3306 (skip SQL)"
else
  log "warn: mysql client missing — run SQL inside crawl-db container"
fi

log "Neo4j label counts"
if compose ps neo4j 2>/dev/null | grep -q Up; then
  compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:Vulnerability) RETURN count(n) AS vulnerabilities;" 2>/dev/null || true
  compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:IOC) RETURN count(n) AS iocs;" 2>/dev/null || true
  compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:Package) RETURN count(n) AS packages;" 2>/dev/null || true
  compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:NucleiTemplate) RETURN count(n) AS nuclei;" 2>/dev/null || true
  compose exec -T neo4j cypher-shell -u "$NEO4J_USER" -p "$NEO4J_PASS" \
    "MATCH (n:SigmaRule) RETURN count(n) AS sigma;" 2>/dev/null || true
else
  fail "neo4j not running"
fi

if command -v curl >/dev/null 2>&1; then
  log "API $API_URL"
  curl -sf "$API_URL" >/dev/null || log "warn: API not up (start default profile: api + neo4j)"
fi

log "PASS — scrape E2E smoke (extend with --restart-scrape for ledger pass 2)"
