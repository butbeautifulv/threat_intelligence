#!/usr/bin/env bash
# Build a versioned ZIP "graph pack" (manifest.json + graph.cypher) for GitHub Releases.
# Usage:
#   ./scripts/build-graph-pack.sh [graphVersion]
# Env:
#   EXPORT_FIRST=1     — run ./scripts/export-graph-cypher.sh before packing (needs running neo4j).
#   GRAPH_PACK_VERSION — overrides optional positional graphVersion.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${GRAPH_PACK_VERSION:-${1:-}}"
if [[ -z "${VERSION}" ]]; then
  VERSION="$(date -u +%Y%m%dT%H%MZ)"
fi

OUT_DIR="${ROOT}/data/neo4j_export/releases"
CY="${ROOT}/data/neo4j_export/graph.cypher"
mkdir -p "${OUT_DIR}"

if [[ "${EXPORT_FIRST:-}" == "1" ]]; then
  "${ROOT}/scripts/export-graph-cypher.sh"
fi

if [[ ! -f "${CY}" ]]; then
  echo "Missing ${CY}. Run: ./scripts/export-graph-cypher.sh (Neo4j must be up), or EXPORT_FIRST=1 ./scripts/build-graph-pack.sh" >&2
  exit 1
fi

sha256_file() {
  local f="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$f" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$f" | awk '{print $1}'
  else
    openssl dgst -sha256 "$f" | awk '{print $2}'
  fi
}

SHA="$(sha256_file "${CY}")"
CREATED="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

cp "${CY}" "${TMP}/graph.cypher"
export VERSION CREATED SHA
python3 - <<'PY' > "${TMP}/manifest.json"
import json, os
m = {
    "schema": "threat-intelligence.graph-pack/1",
    "graphVersion": os.environ["VERSION"],
    "createdAt": os.environ["CREATED"],
    "cypherFile": "graph.cypher",
    "sha256": os.environ["SHA"],
    "neo4j": "5.x",
    "notes": "Import with scripts/import-graph-pack.sh. Plain Cypher from APOC apoc.export.cypher.all; target Neo4j 5.x.",
}
print(json.dumps(m, indent=2))
PY

ZIP="${OUT_DIR}/threat-intel-graph-${VERSION}.zip"
( cd "${TMP}" && zip -q "${ZIP}" manifest.json graph.cypher )
cp "${TMP}/manifest.json" "${OUT_DIR}/manifest.${VERSION}.json"
echo "Pack: ${ZIP}"
echo "Manifest copy: ${OUT_DIR}/manifest.${VERSION}.json"
echo "sha256(graph.cypher)= ${SHA}"
