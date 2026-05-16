#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MCP="${ROOT}/.external/hexstrike-ai-master/hexstrike_mcp.py"
CATALOG="${ROOT}/engage/serve/catalog/tools.yaml"

if [[ ! -f "${MCP}" ]]; then
  echo "skip parity: missing ${MCP}" >&2
  exit 0
fi

legacy=$(grep -cE '@mcp\.tool' "${MCP}" || true)
yaml=$(grep -cE '^  - name:' "${CATALOG}" || true)

echo "legacy mcp tools: ${legacy}"
echo "catalog tools: ${yaml}"

if [[ "${yaml}" -lt 150 ]]; then
  echo "catalog count < 150" >&2
  exit 1
fi

python3 - "${MCP}" "${CATALOG}" <<'PY'
import re, sys
mcp_path, cat_path = sys.argv[1], sys.argv[2]
text = open(mcp_path, encoding="utf-8", errors="replace").read()
legacy = set(re.findall(r"@mcp\.tool\(\)\s*\n\s*def\s+([a-zA-Z0-9_]+)\s*\(", text))
cat = set(re.findall(r"^  - name: (\S+)", open(cat_path, encoding="utf-8").read(), re.M))
missing = sorted(legacy - cat)
extra = sorted(cat - legacy)
if missing:
    print("missing in catalog:", missing[:10], "..." if len(missing) > 10 else "")
    sys.exit(1)
if extra:
    print("extra in catalog:", extra[:5])
bridge = {
    "analyze_target_intelligence", "create_attack_chain_ai", "intelligent_smart_scan",
    "comprehensive_api_audit", "correlate_threat_intelligence", "discover_attack_chains",
    "ai_vulnerability_assessment", "vulnerability_intelligence_dashboard",
    "target_timeline_intelligence",
    "ctf_create_challenge_workflow", "ctf_auto_solve_challenge", "ctf_suggest_tools",
    "ctf_team_strategy", "ctf_cryptography_solver", "ctf_forensics_analyzer", "ctf_binary_analyzer",
}
missing_bridge = sorted(bridge - cat)
if missing_bridge:
    print("intelligence bridge tools missing from catalog:", missing_bridge)
    sys.exit(1)
print("parity OK:", len(legacy), "tools")
PY
