#!/usr/bin/env bash
# Best-effort CI matrix: smoke catalog tools from effectiveness tier (score >= 0.85).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SMOKE="${ROOT}/scripts/test/smoke-engage-tool.sh"
GEN="${ROOT}/scripts/engage/tool-matrix-from-effectiveness.py"
TARGETS="${ROOT}/scripts/engage/tool-matrix.targets"
chmod +x "${SMOKE}" "${GEN}" 2>/dev/null || true

python3 "${GEN}" 0.85

if [[ ! -f "${TARGETS}" ]]; then
  echo "skip tool matrix: no targets file" >&2
  exit 0
fi

mapfile -t tools < "${TARGETS}"
ran=0
for entry in "${tools[@]}"; do
  [[ -z "${entry}" ]] && continue
  tool="${entry%%:*}"
  target="${entry#*:}"
  # Resolve binary from catalog name prefix.
  bin="${tool%%_*}"
  if [[ "${tool}" == *"_"* ]]; then
    bin="${tool%%_*}"
  fi
  case "${tool}" in
    httpx_*) bin=httpx ;;
    nuclei_*) bin=nuclei ;;
    subfinder_*) bin=subfinder ;;
    feroxbuster_*) bin=feroxbuster ;;
    rustscan_*) bin=rustscan ;;
    dalfox_*) bin=dalfox ;;
    enum4linux_*) bin=enum4linux ;;
    waybackurls_*) bin=waybackurls ;;
    masscan_*) bin=masscan ;;
    paramspider_*) bin=paramspider ;;
    katana_*) bin=katana ;;
  esac
  if ! command -v "${bin}" >/dev/null 2>&1; then
    echo "skip ${tool}: ${bin} not on PATH" >&2
    continue
  fi
  echo "smoke ${tool} -> ${target}"
  "${SMOKE}" "${tool}" "${target}" || true
  ran=$((ran + 1))
done

min="${ENGAGE_TOOL_MATRIX_MIN:-15}"
strict="${ENGAGE_TOOL_MATRIX_STRICT:-0}"
if [[ "${ran}" -eq 0 ]]; then
  echo "skip tool matrix: no supported binaries on PATH" >&2
  exit 0
fi
if [[ "${strict}" == "1" && "${ran}" -lt 30 ]]; then
  echo "FAIL: strict mode requires >=30 tools, got ${ran}" >&2
  exit 1
fi
if [[ "${ran}" -lt "${min}" ]]; then
  echo "WARN: matrix ran ${ran} tools (min ${min}); best-effort skip" >&2
fi
echo "OK engage tool matrix (${ran}/${#tools[@]} exercised; skips when binary missing)"
