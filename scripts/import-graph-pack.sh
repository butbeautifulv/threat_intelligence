#!/usr/bin/env bash
# Import a graph pack ZIP (from GitHub Releases or local path): verify manifest sha256, stream Cypher to Neo4j.
# Usage:
#   ./scripts/import-graph-pack.sh <https://.../threat-intel-graph-X.zip | /path/to/pack.zip>
# Env:
#   NEO4J_URI   (default neo4j://localhost:7687)
#   NEO4J_USER  (default neo4j)
#   NEO4J_PASS
#   NEO4J_DB    (default neo4j)
#   USE_DOCKER_COMPOSE=1 — run cypher-shell via `docker compose exec` from repo root (uses compose service `neo4j`).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SOURCE="${1:?usage: $0 <https://.../*.zip | /path/to/*.zip>}"

COMPOSE="${COMPOSE:-docker compose}"
COMPOSE_FILES="${COMPOSE_FILES:--f deploy/scrape/compose.yml -f deploy/pipeline/compose.yml -f deploy/graph/compose.yml}"

NEO4J_URI="${NEO4J_URI:-neo4j://localhost:7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASS="${NEO4J_PASS:-neo4jpassword}"
NEO4J_DB="${NEO4J_DB:-neo4j}"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

if [[ "${SOURCE}" == https://* ]] || [[ "${SOURCE}" == http://* ]]; then
  echo "Downloading pack..."
  curl -fsSL "${SOURCE}" -o "${TMP}/pack.zip"
else
  cp "${SOURCE}" "${TMP}/pack.zip"
fi

unzip -q "${TMP}/pack.zip" -d "${TMP}/ex"
M="${TMP}/ex/manifest.json"
CY="${TMP}/ex/graph.cypher"
if [[ ! -f "${M}" ]] || [[ ! -f "${CY}" ]]; then
  echo "ZIP must contain manifest.json and graph.cypher at archive root" >&2
  exit 1
fi

EXP="$(python3 -c "import json; print(json.load(open('${M}'))['sha256'])")"
GOT="$(python3 -c "import hashlib; print(hashlib.sha256(open('${CY}','rb').read()).hexdigest())")"
if [[ "${EXP}" != "${GOT}" ]]; then
  echo "sha256 mismatch: manifest has ${EXP}, file has ${GOT}" >&2
  exit 1
fi
echo "Checksum OK (graphVersion=$(python3 -c "import json; print(json.load(open('${M}'))['graphVersion'])"))"

compose() {
  # shellcheck disable=SC2086
  (cd "${ROOT}" && $COMPOSE $COMPOSE_FILES "$@")
}

run_shell() {
  if [[ "${USE_DOCKER_COMPOSE:-}" == "1" ]]; then
    compose exec -T neo4j cypher-shell \
      -a "${NEO4J_URI}" -u "${NEO4J_USER}" -p "${NEO4J_PASS}" -d "${NEO4J_DB}" "$@"
  else
    if command -v cypher-shell >/dev/null 2>&1; then
      cypher-shell -a "${NEO4J_URI}" -u "${NEO4J_USER}" -p "${NEO4J_PASS}" -d "${NEO4J_DB}" "$@"
    else
      echo "cypher-shell not in PATH; set USE_DOCKER_COMPOSE=1 from repo root or install Neo4j tools" >&2
      exit 1
    fi
  fi
}

echo "Streaming Cypher (this may take a while)..."
# Many statements; stdin is the whole file
run_shell < "${CY}"
echo "Import finished."
