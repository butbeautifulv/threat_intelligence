#!/usr/bin/env python3
"""Generate tools.live.yaml — one overlay row per catalog name (no synthetics).

Subprocess catalog entries: enabled: true (runner dispatch).
bridge_api (binary api or WORKFLOW_BINARIES): enabled: false (dispatch via Get, not MustGet).
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
OUT = ROOT / "engage/serve/catalog/tools.live.yaml"

# Sync with generate-tools-na-matrix.py WORKFLOW_BINARIES.
WORKFLOW_BINARIES = frozenset({
    "api", "bugbounty", "ai", "get", "http", "create", "execute", "generate",
    "list", "kube", "browser", "autorecon", "comprehensive", "advanced",
    "analyze", "checkov", "clair", "cloudmapper", "checksec", "clear",
})


def parse_blocks(text: str) -> list[tuple[str, str, str]]:
    blocks = re.split(r"(?=^  - name: )", text, flags=re.M)
    out: list[tuple[str, str, str]] = []
    for block in blocks[1:]:
        name_m = re.search(r"^\s*- name: (\S+)", block, re.M)
        bin_m = re.search(r"^\s*binary: (\S+)", block, re.M)
        if name_m and bin_m:
            out.append((name_m.group(1), bin_m.group(1), block))
    return out


def is_bridge_api(binary: str) -> bool:
    return binary in WORKFLOW_BINARIES


def set_enabled(block: str, enabled: bool) -> str:
    val = "true" if enabled else "false"
    if re.search(r"^\s*enabled:\s*\S+", block, flags=re.M):
        return re.sub(
            r"^(\s*)enabled:\s*\S+",
            rf"\1enabled: {val}",
            block,
            count=1,
            flags=re.M,
        )
    return block.rstrip() + f"\n    enabled: {val}\n"


def main() -> int:
    if not CATALOG.is_file():
        print(f"catalog missing: {CATALOG}", file=sys.stderr)
        return 1

    text = CATALOG.read_text(encoding="utf-8")
    blocks = parse_blocks(text)
    if len(blocks) != 158:
        print(f"FAIL: expected 158 catalog tools, got {len(blocks)}", file=sys.stderr)
        return 1

    subprocess = 0
    bridge = 0
    out_blocks: list[str] = []
    for name, binary, block in sorted(blocks, key=lambda x: x[0]):
        if is_bridge_api(binary):
            bridge += 1
            out_blocks.append(set_enabled(block, False).rstrip("\n"))
        else:
            subprocess += 1
            out_blocks.append(set_enabled(block, True).rstrip("\n"))

    lines = [
        "# P9h: one overlay per catalog name (158). Subprocess enabled; bridge_api disabled.",
        "# Regenerate: python3 scripts/engage/generate-tools-catalog-live.py",
        "tools:",
    ]
    lines.extend(out_blocks)
    OUT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(
        f"wrote {len(out_blocks)} tools to {OUT} "
        f"({subprocess} subprocess enabled, {bridge} bridge_api disabled)"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
