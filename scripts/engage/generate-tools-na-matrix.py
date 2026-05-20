#!/usr/bin/env python3
"""Generate docs/engage/engage-tools-na-matrix.md — execution status for every catalog tool."""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
LIVE = ROOT / "engage/serve/catalog/tools.live.yaml"
OUT = ROOT / "docs/engage/engage-tools-na-matrix.md"

sys.path.insert(0, str(Path(__file__).resolve().parent))
from runner_binaries import RUNNER_BINARIES  # noqa: E402

WORKFLOW_BINARIES = frozenset({
    "api", "bugbounty", "ai", "get", "http", "create", "execute", "generate",
    "list", "kube", "browser", "autorecon", "comprehensive", "advanced",
    "analyze", "checkov", "clair", "cloudmapper", "checksec", "clear",
})
# P10b: runner subprocess wrappers — not bridge workflow placeholders.
SUBPROCESS_BINARIES = frozenset({"engage-python-exec"})


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


def is_bridge_api(binary: str) -> bool:
    if binary in SUBPROCESS_BINARIES:
        return False
    return binary == "api" or binary in WORKFLOW_BINARIES


def classify(name: str, binary: str, live_enabled: bool) -> tuple[str, str]:
    if is_bridge_api(binary):
        if binary == "api":
            return "bridge_api", "in-process MCP bridge handler"
        return "bridge_api", f"workflow placeholder binary `{binary}`"
    if live_enabled:
        return "live", "enabled in tools.live.yaml (subprocess)"
    if binary in RUNNER_BINARIES:
        return "runner_N/A", "binary in runner image but not enabled in lab profile"
    return "runner_N/A", "binary not in engage-runner image"


def render_md(
    rows: list[tuple[str, str, str, str, str]],
    live_rows: int,
    live_count: int,
    catalog_live: int,
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
        f"**Catalog tools:** {len(rows)} | **Live rows:** {live_rows} | "
        f"**Live enabled (subprocess):** {live_count} | **Catalog ∩ live enabled:** {catalog_live}",
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
    catalog_names = set(catalog)
    live_names = set(live)
    subprocess_catalog = {
        n for n, t in catalog.items() if not is_bridge_api(t.get("binary", ""))
    }

    rows: list[tuple[str, str, str, str, str]] = []
    for name in sorted(catalog):
        meta = catalog[name]
        binary = meta.get("binary", "")
        category = meta.get("category", "")
        in_live = name in live_enabled
        status, reason = classify(name, binary, in_live)
        rows.append((name, binary, category, status, reason))

    live_rows = len(live)
    live_count = len(live_enabled)
    catalog_live = sum(1 for r in rows if r[3] == "live")
    permanent_na = sum(1 for r in rows if r[3] == "permanent_N/A")
    runner_na = sum(1 for r in rows if r[3] == "runner_N/A")
    orphans = live_names - catalog_names
    body = render_md(rows, live_rows, live_count, catalog_live)
    OUT.write_text(body, encoding="utf-8")
    print(
        f"wrote {OUT} ({len(rows)} catalog, {live_rows} live rows, {live_count} live enabled, "
        f"{catalog_live} catalog live, runner_N/A={runner_na}, permanent_N/A={permanent_na})"
    )

    if len(rows) != 158:
        print(f"FAIL: expected 158 catalog tools, got {len(rows)}", file=sys.stderr)
        return 1
    if live_rows != 158:
        print(f"FAIL: expected 158 live overlay rows, got {live_rows}", file=sys.stderr)
        return 1
    if orphans:
        print(f"FAIL: {len(orphans)} live names not in catalog: {sorted(orphans)[:5]}...", file=sys.stderr)
        return 1
    if live_names != catalog_names:
        missing = catalog_names - live_names
        print(f"FAIL: {len(missing)} catalog names missing from live", file=sys.stderr)
        return 1
    if live_count != len(subprocess_catalog):
        print(
            f"FAIL: live enabled {live_count} != subprocess catalog {len(subprocess_catalog)}",
            file=sys.stderr,
        )
        return 1
    if live_count < args.min_live:
        print(f"FAIL: live count {live_count} < {args.min_live}", file=sys.stderr)
        return 1
    if args.check and permanent_na > 0:
        print(f"FAIL: permanent_N/A count {permanent_na} (full port requires 0)", file=sys.stderr)
        return 1
    if args.check and runner_na > 0:
        print(f"FAIL: runner_N/A count {runner_na} (P9i requires 0)", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
