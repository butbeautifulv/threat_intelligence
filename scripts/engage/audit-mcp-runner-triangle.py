#!/usr/bin/env python3
"""MCP ↔ catalog ↔ runner triangle for engage audit report."""
from __future__ import annotations

import csv
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MCP = ROOT / ".external/hexstrike-ai-master/hexstrike_mcp.py"
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
LIVE = ROOT / "engage/serve/catalog/tools.live.yaml"
OUT = ROOT / "docs/engage/engage-mcp-runner-triangle.csv"
LIST_BIN = ROOT / "scripts/engage/list-runner-binaries.sh"


def runner_bins() -> set[str]:
    if not LIST_BIN.is_file():
        return set()
    try:
        out = subprocess.check_output(["bash", str(LIST_BIN)], text=True, timeout=60)
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
        return set()
    return {ln.strip() for ln in out.splitlines() if ln.strip()}


def parse_yaml_tools(path: Path) -> dict[str, dict]:
    text = path.read_text(encoding="utf-8")
    tools: dict[str, dict] = {}
    blocks = re.split(r"(?=^  - name: )", text, flags=re.M)
    for block in blocks[1:]:
        name_m = re.search(r"^\s*- name: (\S+)", block, re.M)
        if not name_m:
            continue
        name = name_m.group(1)
        bin_m = re.search(r"^\s*binary: (\S+)", block, re.M)
        en_m = re.search(r"^\s*enabled:\s*(\S+)", block, re.M)
        tools[name] = {
            "binary": bin_m.group(1) if bin_m else "",
            "enabled": en_m.group(1) == "true" if en_m else False,
        }
    return tools


def mcp_names() -> set[str]:
    text = MCP.read_text(encoding="utf-8", errors="replace")
    legacy = set(re.findall(r"@mcp\.tool\(\)\s*\n\s*def\s+([a-zA-Z0-9_]+)\s*\(", text))
    bridge = {
        "target_timeline_intelligence",
        "ctf_create_challenge_workflow", "ctf_auto_solve_challenge", "ctf_suggest_tools",
        "ctf_team_strategy", "ctf_cryptography_solver", "ctf_forensics_analyzer", "ctf_binary_analyzer",
    }
    return legacy | bridge


def main() -> int:
    if not CATALOG.is_file():
        print(f"missing {CATALOG}", file=sys.stderr)
        return 1
    catalog = parse_yaml_tools(CATALOG)
    live = parse_yaml_tools(LIVE) if LIVE.is_file() else {}
    enabled = {n for n, t in live.items() if t.get("enabled")} or {
        n for n, t in catalog.items() if t.get("enabled")
    }
    bins = runner_bins()
    mcp = mcp_names()

    rows: list[list[str]] = []
    for name in sorted(enabled):
        binary = live.get(name, catalog.get(name, {})).get("binary", "")
        runnable = "yes" if binary == "api" or binary in bins else "no"
        in_mcp = "yes" if name in mcp else "bridge"
        notes = []
        if binary == "api":
            notes.append("in-process bridge")
        elif runnable == "no":
            notes.append("binary not in runner image")
        rows.append([name, binary, runnable, in_mcp, ";".join(notes)])

    OUT.parent.mkdir(parents=True, exist_ok=True)
    with OUT.open("w", newline="", encoding="utf-8") as f:
        w = csv.writer(f)
        w.writerow(["enabled", "binary", "runnable", "in_mcp", "notes"])
        w.writerows(rows)

    runnable_n = sum(1 for r in rows if r[2] == "yes")
    print(f"enabled tools: {len(rows)}")
    print(f"runnable in runner: {runnable_n}/{len(rows)}")
    print(f"wrote {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
