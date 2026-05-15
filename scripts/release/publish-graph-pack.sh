#!/usr/bin/env bash
# Build graph pack ZIP and create GitHub release veil-graph-vX.Y.Z.
# Usage: GRAPH_PACK_VERSION=v0.4.0 ./scripts/release/publish-graph-pack.sh [--draft] [--skip-build]
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

DRAFT=""
SKIP_BUILD=0
for arg in "$@"; do
  case "$arg" in
    --draft) DRAFT="--draft" ;;
    --skip-build) SKIP_BUILD=1 ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

VERSION="${GRAPH_PACK_VERSION:?set GRAPH_PACK_VERSION e.g. v0.4.0}"
VERSION="$(pack_normalize_version "${VERSION}")"
TAG="$(pack_release_tag "${VERSION}")"
ZIP="${PACK_RELEASES_DIR}/$(pack_zip_name "${VERSION}")"

if [[ "$SKIP_BUILD" -eq 0 ]]; then
  EXPORT_FIRST=1 "${VEIL_ROOT}/scripts/graph-pack/build.sh" "${VERSION}"
fi

if [[ ! -f "${ZIP}" ]]; then
  echo "missing pack: ${ZIP}" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI required for publish" >&2
  exit 1
fi

NOTES="Graph pack ${VERSION} for Veil (Neo4j 5.x). Import via scripts/graph-pack/import.sh or graph-bootstrap."
gh_args=(release create "${TAG}" --repo butbeautifulv/veil --title "${TAG}" --notes "${NOTES}")
[[ -n "${DRAFT}" ]] && gh_args+=(--draft)
gh "${gh_args[@]}" "${ZIP}"

echo "Published: ${TAG}"
echo "GRAPH_PACK_DEFAULT_URL=$(pack_release_url "${VERSION}")"
