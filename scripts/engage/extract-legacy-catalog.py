#!/usr/bin/env python3
"""Extract MCP tool names and parameters from legacy reference into engage catalog YAML."""
from __future__ import annotations

import ast
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MCP = ROOT / ".external/hexstrike-ai-master/hexstrike_mcp.py"
OUT = ROOT / "engage/serve/catalog/tools.yaml"

PREFIX_CAT = [
    ("nmap", "network"), ("rustscan", "network"), ("masscan", "network"),
    ("gobuster", "web"), ("nuclei", "web"), ("httpx", "web"), ("ffuf", "web"),
    ("nikto", "web"), ("sqlmap", "web"), ("wpscan", "web"),
    ("prowler", "cloud"), ("trivy", "cloud"), ("scout", "cloud"), ("kube", "cloud"),
    ("hydra", "auth"), ("hashcat", "auth"), ("john", "auth"),
    ("amass", "osint"), ("subfinder", "osint"), ("theharvester", "osint"),
    ("gdb", "binary"), ("ghidra", "binary"), ("radare", "binary"), ("binwalk", "binary"),
]

DEFAULT_BINARY = {
    "network": "nmap", "web": "httpx", "cloud": "trivy", "auth": "hydra",
    "osint": "subfinder", "binary": "strings", "ctf": "file", "intelligence": "echo",
}

# Per-tool arg templates when parameters imply structured CLI.
ARGS_TEMPLATES: dict[str, list[str]] = {
    "nmap_scan": ["{scan_type}", "-p", "{ports}", "{additional_args}", "{target}"],
    "nuclei_scan": ["-u", "{target}", "-t", "{templates}", "{additional_args}"],
    "httpx_probe": ["-u", "{target}", "{additional_args}"],
    "subfinder_scan": ["-d", "{target}", "{additional_args}"],
    "trivy_scan": ["image", "{target}", "{additional_args}"],
    "gobuster_scan": ["dir", "-u", "{target}", "-w", "{wordlist}", "{additional_args}"],
}


def category_for(name: str) -> str:
    low = name.lower()
    for prefix, cat in PREFIX_CAT:
        if low.startswith(prefix) or prefix in low:
            return cat
    if "cloud" in low or "aws" in low:
        return "cloud"
    if "ctf" in low:
        return "ctf"
    if "bugbounty" in low or "intelligence" in low or "analyze" in low:
        return "intelligence"
    return "web"


def describe(name: str, cat: str) -> str:
    return f"{cat} tool: {name.replace('_', ' ')}"


def parse_mcp_tools(text: str) -> dict[str, list[dict]]:
    """Parse @mcp.tool function signatures into parameter metadata."""
    tools: dict[str, list[dict]] = {}
    pattern = re.compile(
        r"@mcp\.tool\(\)\s*\n\s*def\s+([a-zA-Z0-9_]+)\s*\(([^)]*)\)",
        re.MULTILINE,
    )
    for match in pattern.finditer(text):
        name = match.group(1)
        params_src = match.group(2).strip()
        if not params_src:
            tools[name] = [{"name": "target", "type": "string", "required": True}]
            continue
        try:
            fake = f"def _f({params_src}): pass"
            tree = ast.parse(fake)
            fn = tree.body[0]
            assert isinstance(fn, ast.FunctionDef)
            plist = []
            for arg in fn.args.args:
                pname = arg.arg
                if pname in ("self", "cls"):
                    continue
                entry = {"name": pname, "type": "string", "required": True}
                plist.append(entry)
            defaults = [None] * (len(fn.args.args) - len(fn.args.defaults)) + list(fn.args.defaults)
            for i, d in enumerate(defaults):
                if d is None:
                    continue
                if isinstance(d, ast.Constant):
                    plist[i]["default"] = str(d.value)
                    plist[i]["required"] = False
            if not any(p["name"] == "target" for p in plist):
                plist.insert(0, {"name": "target", "type": "string", "required": True})
            tools[name] = plist
        except SyntaxError:
            tools[name] = [{"name": "target", "type": "string", "required": True}]
    return tools


def yaml_quote(s: str) -> str:
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'


def main() -> int:
    if not MCP.is_file():
        print(f"missing {MCP}", file=sys.stderr)
        return 1
    text = MCP.read_text(encoding="utf-8", errors="replace")
    param_map = parse_mcp_tools(text)
    names = sorted(param_map.keys())

    lines = [
        "# Veil engage tool catalog (names aligned with legacy MCP reference in .external/)",
        "# Regenerate: make catalog-engage",
        "# enabled=false until runner image provides the binary on PATH.",
        "tools:",
    ]
    for name in names:
        cat = category_for(name)
        binary = name.split("_")[0] if "_" in name else name
        if len(binary) > 20:
            binary = DEFAULT_BINARY.get(cat, "echo")
        params = param_map.get(name, [{"name": "target", "type": "string", "required": True}])
        args = ARGS_TEMPLATES.get(name, ["{target}", "{additional_args}"])

        lines.append(f"  - name: {name}")
        lines.append(f"    category: {cat}")
        lines.append(f"    binary: {binary}")
        lines.append("    parameters:")
        for p in params:
            lines.append(f"      - name: {p['name']}")
            lines.append(f"        type: {p.get('type', 'string')}")
            if p.get("default") is not None:
                lines.append(f"        default: {yaml_quote(str(p['default']))}")
            if not p.get("required", True):
                lines.append("        required: false")
            else:
                lines.append("        required: true")
        lines.append("    args:")
        for a in args:
            lines.append(f"      - {yaml_quote(a)}")
        lines.append(f"    timeout_sec: 300")
        lines.append(f"    description: {yaml_quote(describe(name, cat))}")
        lines.append(f"    enabled: false")

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"wrote {len(names)} tools to {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
