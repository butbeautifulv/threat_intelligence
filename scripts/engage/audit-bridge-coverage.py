#!/usr/bin/env python3
"""Audit bridge_api catalog tools — each must map to a bridge handler path."""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
LIVE = ROOT / "engage/serve/catalog/tools.live.yaml"
DISPATCH = ROOT / "engage/serve/internal/usecase/tooldispatch"

WORKFLOW_BINARIES = frozenset({
    "api", "bugbounty", "ai", "get", "http", "create", "execute", "generate",
    "list", "kube", "browser", "autorecon", "comprehensive", "advanced",
    "analyze", "checkov", "clair", "cloudmapper", "checksec", "clear",
})
SUBPROCESS_BINARIES = frozenset({"engage-python-exec"})

IS_INTEL_BRIDGE_BY_NAME = frozenset({
    "comprehensive_api_audit",
    "target_timeline_intelligence",
    "target_graph_context",
    "monitor_cve_feeds",
    "generate_exploit_from_cve",
})


def parse_tools(path: Path) -> dict[str, dict[str, str]]:
    text = path.read_text(encoding="utf-8")
    tools: dict[str, dict[str, str]] = {}
    for block in re.split(r"(?=^  - name: )", text, flags=re.M)[1:]:
        name_m = re.search(r"^\s*- name: (\S+)", block, re.M)
        if not name_m:
            continue
        name = name_m.group(1)
        bin_m = re.search(r"^\s*binary: (\S+)", block, re.M)
        cat_m = re.search(r"^\s*category: (\S+)", block, re.M)
        tools[name] = {
            "binary": bin_m.group(1) if bin_m else "",
            "category": cat_m.group(1) if cat_m else "",
        }
    return tools


def is_bridge_api(name: str, binary: str, category: str) -> bool:
    if binary in SUBPROCESS_BINARIES:
        return False
    if binary == "api":
        return True
    if binary in WORKFLOW_BINARIES:
        return True
    return False


def map_keys(path: Path, var_name: str) -> set[str]:
    text = path.read_text(encoding="utf-8")
    m = re.search(rf"var {var_name}\s*=\s*map\[string\].*?\{{(.*?)\n\}}", text, re.S)
    if not m:
        return set()
    body = m.group(1)
    return set(re.findall(r'"([^"]+)":', body))


def agent_tool_cases(path: Path) -> set[str]:
    text = path.read_text(encoding="utf-8")
    names: set[str] = set()
    for block in re.findall(r'case\s+((?:"[^"]+"\s*,\s*)*"[^"]+")', text):
        names.update(re.findall(r'"([^"]+)"', block))
    if 'strings.HasPrefix(name, "ai_generate_")' in text:
        names.add("ai_generate_*")
    if "ai_reconnaissance_workflow" in text:
        names.add("ai_reconnaissance_workflow")
    if "ai_test_payload" in text:
        names.add("ai_test_payload")
    return names


def playbook_names() -> set[str]:
    names: set[str] = set()
    for path in (ROOT / "engage/serve/playbooks").glob("*.yaml"):
        for m in re.finditer(r"^\s*- name:\s*(\S+)", path.read_text(encoding="utf-8"), re.M):
            names.add(m.group(1))
    return names


def coverage_for(
    name: str,
    binary: str,
    category: str,
    intel: set[str],
    ctf: set[str],
    cve: set[str],
    agent: set[str],
    workflow: set[str],
    playbooks: set[str],
) -> str | None:
    if name in intel:
        return "intel_bridge_handlers"
    if name in ctf or category == "ctf" or name.startswith("ctf_"):
        if name in ctf:
            return "ctf_bridge_handlers"
        return None
    if name in cve or name in IS_INTEL_BRIDGE_BY_NAME:
        return "cve_bridge_handlers"
    if name in workflow:
        return "bridge_workflow_handlers"
    if name in agent:
        return "tryAgentTool"
    if name.startswith("ai_generate_"):
        return "tryAgentTool:ai_generate_*"
    if name in playbooks:
        return "tryPlaybookByName"
    if category == "intelligence" or name in IS_INTEL_BRIDGE_BY_NAME:
        return None
    if is_bridge_api(name, binary, category):
        return None
    return "n/a"


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--min-covered", type=int, default=55)
    args = ap.parse_args()

    catalog = parse_tools(CATALOG)
    live = parse_tools(LIVE) if LIVE.is_file() else {}
    live_enabled = {n for n, t in live.items() if t.get("enabled") == "true"}

    intel = map_keys(DISPATCH / "bridge_handlers.go", "intelBridgeHandlers")
    ctf = map_keys(DISPATCH / "bridge_ctf.go", "ctfBridgeHandlers")
    cve = map_keys(DISPATCH / "bridge_cve.go", "cveBridgeHandlers")
    workflow = map_keys(DISPATCH / "bridge_workflow.go", "bridgeWorkflowHandlers")
    agent = agent_tool_cases(DISPATCH / "agent_tools.go")
    playbooks = playbook_names()

    bridge_tools = [
        n
        for n, meta in sorted(catalog.items())
        if is_bridge_api(n, meta.get("binary", ""), meta.get("category", ""))
    ]

    covered: list[tuple[str, str]] = []
    missing: list[str] = []
    for name in bridge_tools:
        src = coverage_for(
            name,
            catalog[name].get("binary", ""),
            catalog[name].get("category", ""),
            intel,
            ctf,
            cve,
            agent,
            workflow,
            playbooks,
        )
        if src:
            covered.append((name, src))
        else:
            missing.append(name)

    total = len(bridge_tools)
    n_cov = len(covered)
    print(f"bridge_api tools: {total}")
    print(f"covered: {n_cov}/{total}")
    if missing:
        print("missing:", ", ".join(missing), file=sys.stderr)
        for name in missing:
            meta = catalog[name]
            print(
                f"  - {name} binary={meta.get('binary')} category={meta.get('category')}",
                file=sys.stderr,
            )
    for name, src in covered:
        print(f"  ok {name} -> {src}")

    if n_cov < args.min_covered:
        print(
            f"FAIL: bridge coverage {n_cov}/{total} < --min-covered {args.min_covered}",
            file=sys.stderr,
        )
        return 1
    if missing:
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
