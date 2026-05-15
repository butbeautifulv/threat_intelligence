#!/usr/bin/env bash
# Push repository description from .github/repo-description.txt to GitHub.
set -euo pipefail
# shellcheck source=../lib/common.sh
source "$(cd "$(dirname "$0")/.." && pwd)/lib/common.sh"

DESC_FILE="${VEIL_ROOT}/.github/repo-description.txt"
if [[ ! -f "${DESC_FILE}" ]]; then
  echo "missing ${DESC_FILE}" >&2
  exit 1
fi
if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI required" >&2
  exit 1
fi

DESC="$(tr -d '\n' <"${DESC_FILE}")"
gh repo edit butbeautifulv/veil --description "${DESC}"
echo "Updated GitHub description for butbeautifulv/veil"
