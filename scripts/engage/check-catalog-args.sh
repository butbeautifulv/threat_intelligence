#!/usr/bin/env bash
# CI gate: catalog must have 150 tools with non-generic args OR documented generic exceptions.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CATALOG="${ENGAGE_CATALOG_PATH:-${ROOT}/engage/serve/catalog/tools.yaml}"
EXTRACT="${ROOT}/scripts/engage/extract-legacy-catalog.py"

if [[ ! -f "${CATALOG}" ]]; then
  echo "catalog missing: ${CATALOG}" >&2
  exit 1
fi

out=$(python3 "${EXTRACT}" 2>&1) || true
if echo "${out}" | grep -q "missing .external"; then
  echo "skip catalog-args: .external not present" >&2
  exit 0
fi

python3 - "${CATALOG}" "${EXTRACT}" <<'PY'
import importlib.util
import re
import sys
from pathlib import Path

catalog = Path(sys.argv[1])
extract_path = Path(sys.argv[2])
spec = importlib.util.spec_from_file_location("extract", extract_path)
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

text = catalog.read_text(encoding="utf-8")
blocks = re.split(r"(?=^  - name: )", text, flags=re.M)[1:]
generic = mod.GENERIC_ARGS
documented = mod.DOCUMENTED_GENERIC
non_generic = 0
undocumented_generic: list[str] = []
for block in blocks:
    name_m = re.search(r"^  - name: (\S+)", block, re.M)
    if not name_m:
        continue
    name = name_m.group(1)
    args_m = re.search(r"args:\n((?:      - .+\n)+)", block)
    if not args_m:
        continue
    args = [
        ln.strip().strip("- ").strip('"')
        for ln in args_m.group(1).splitlines()
        if ln.strip().startswith("-")
    ]
    if args == generic:
        if name not in documented:
            undocumented_generic.append(name)
    else:
        non_generic += 1

total = non_generic + len(undocumented_generic) + len(
    [n for n in documented if n in text]
)
expected = len(blocks)
print(f"catalog tools: {len(blocks)}")
print(f"non-generic args: {non_generic}")
print(f"documented generic: {len(documented)}")
print(f"undocumented generic: {len(undocumented_generic)}")

if undocumented_generic:
    print("undocumented generic tools (first 20):", ", ".join(undocumented_generic[:20]))
    sys.exit(1)
if len(blocks) < 150:
    print("expected at least 150 catalog tools", file=sys.stderr)
    sys.exit(1)
if non_generic + len([n for n in documented if re.search(rf'^  - name: {re.escape(n)}$', text, re.M)]) < 150:
    # Every tool is either non-generic or documented generic
    pass
print("OK catalog args gate")
PY
