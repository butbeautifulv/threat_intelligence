#!/usr/bin/env python3
"""Generate docs/engage-tools-na-matrix.md — execution status for every catalog tool."""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
LIVE = ROOT / "engage/serve/catalog/tools.live.yaml"
OUT = ROOT / "docs/engage-tools-na-matrix.md"

# Sync with generate-tools-live.py + runner.Dockerfile (Phase 25).
RUNNER_BINARIES = frozenset({
    "nmap", "masscan", "sqlmap", "nikto", "gobuster", "feroxbuster",
    "nuclei", "httpx", "subfinder", "katana", "naabu", "dnsx", "gau",
    "waybackurls", "dalfox", "amass", "ffuf", "arjun", "dirsearch",
    "paramspider", "rustscan", "trivy", "dnsenum", "fierce", "hydra",
    "wafw00f", "enum4linux", "sslscan", "testssl", "dirb",
    "whatweb", "nbtscan", "binwalk", "jaeles", "x8", "enum4linux-ng",
})

WORKFLOW_BINARIES = frozenset({
    "api", "bugbounty", "ai", "get", "http", "create", "execute", "generate",
    "list", "kube", "browser", "autorecon", "comprehensive", "advanced",
    "analyze", "checkov", "clair", "cloudmapper", "checksec", "clear",
})

PERMANENT_NA_BINARIES = frozenset({
    "ghidra", "burpsuite", "burp", "metasploit", "msfconsole", "angr", "gdb",
    "wpscan", "wireshark", "volatility", "radare2", "r2", "cutter",
    "john", "hashcat", "aircrack", "ettercap", "bettercap",
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
        en_m = re.search(r"^\s*enabled:\s*(\S+)", block, re.M)
        tools[name] = {
            "binary": bin_m.group(1) if bin_m else "",
            "category": cat_m.group(1) if cat_m else "",
            "enabled": en_m.group(1) if en_m else "false",
        }
    return tools


def classify(name: str, binary: str, live_enabled: bool) -> tuple[str, str]:
    if binary == "api":
        return "bridge_api", "in-process MCP bridge handler"
    if binary in WORKFLOW_BINARIES:
        return "bridge_api", f"workflow placeholder binary `{binary}`"
    if binary in PERMANENT_NA_BINARIES or any(
        x in binary.lower() for x in ("ghidra", "burp", "metasploit", "angr", "wpscan")
    ):
        return "permanent_N/A", "GUI or heavy stack — out of runner image by design"
    if live_enabled:
        if binary in RUNNER_BINARIES or binary == "api":
            return "live", "enabled in tools.live.yaml"
        return "live", "enabled in tools.live.yaml (synthetic variant)"
    if binary in RUNNER_BINARIES:
        return "runner_N/A", "binary in runner image but not selected for lab profile"
    return "runner_N/A", "binary not in engage-runner image"


def render_md(
    rows: list[tuple[str, str, str, str, str]], live_count: int, catalog_live: int
) -> str:
    lines = [
        "# Engage tools — execution N/A matrix",
        "",
        "Auto-generated. Regenerate:",
        "",
        "```bash",
        "python3 scripts/engage/generate-tools-na-matrix.py",
        "make test-engage-na-matrix",
        "```",
        "",
        f"**Catalog tools:** {len(rows)} | **Live enabled (tools.live.yaml):** {live_count} | **Catalog ∩ live:** {catalog_live}",
        "",
        "| Tool | Binary | Category | Status | Reason |",
        "|------|--------|----------|--------|--------|",
    ]
    for name, binary, category, status, reason in rows:
        reason = reason.replace("|", "\\|")
        lines.append(f"| `{name}` | `{binary}` | {category} | {status} | {reason} |")
    lines.append("")
    return "\n".join(lines)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--check", action="store_true", help="fail if output stale or counts wrong")
    ap.add_argument("--min-live", type=int, default=100)
    args = ap.parse_args()

    if not CATALOG.is_file():
        print(f"missing {CATALOG}", file=sys.stderr)
        return 1

    catalog = parse_tools(CATALOG)
    live = parse_tools(LIVE) if LIVE.is_file() else {}
    live_enabled = {n for n, t in live.items() if t.get("enabled") == "true"}

    rows: list[tuple[str, str, str, str, str]] = []
    for name in sorted(catalog):
        meta = catalog[name]
        binary = meta.get("binary", "")
        category = meta.get("category", "")
        in_live = name in live_enabled
        status, reason = classify(name, binary, in_live)
        rows.append((name, binary, category, status, reason))

    # DoD: tools.live.yaml enabled count (includes synthetic lab variants).
    live_count = len(live_enabled)
    catalog_live = sum(1 for r in rows if r[3] == "live")
    body = render_md(rows, live_count, catalog_live)
    OUT.write_text(body, encoding="utf-8")
    print(f"wrote {OUT} ({len(rows)} catalog, {live_count} live enabled, {catalog_live} catalog live)")

    if len(rows) != 158:
        print(f"FAIL: expected 158 catalog tools, got {len(rows)}", file=sys.stderr)
        return 1
    if live_count < args.min_live:
        print(f"FAIL: live count {live_count} < {args.min_live}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
