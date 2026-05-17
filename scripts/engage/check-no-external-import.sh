#!/usr/bin/env bash
# P11d: engage runtime must not import or embed paths under .external/hexstrike*.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}"

PATTERN='\.external/hexstrike'
DIRS=(engage pkg/engage)

matches=()
for dir in "${DIRS[@]}"; do
  [[ -d "${ROOT}/${dir}" ]] || continue
  while IFS= read -r -d '' f; do
    matches+=("${f}")
  done < <(rg -l --glob '*.go' "${PATTERN}" "${dir}" 2>/dev/null || true)
done

if ((${#matches[@]} > 0)); then
  echo "ERROR: ${DIRS[*]} must not reference ${PATTERN} (archive-only; use engage catalog/parity scripts at repo root)" >&2
  rg -n --glob '*.go' "${PATTERN}" "${DIRS[@]}" >&2 || true
  exit 1
fi

echo "OK: no .external/hexstrike references in engage/ or pkg/engage Go sources"
