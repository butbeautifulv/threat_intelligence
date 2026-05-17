#!/usr/bin/env python3
"""Extract HexStrike tool_effectiveness tables to JSON for Go parity checks."""
from __future__ import annotations

import ast
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SERVER = ROOT / ".external/hexstrike-ai-master/hexstrike_server.py"
OUT_JSON = ROOT / "pkg/decision/testdata/effectiveness_legacy.json"

# HexStrike TargetType enum values -> engage target types
TYPE_MAP = {
    "web_application": "web",
    "network_host": "ip",
    "api_endpoint": "api",
    "cloud_service": "cloud",
    "binary_file": "binary",
}


def main() -> int:
    if not SERVER.is_file():
        print(f"skip: missing {SERVER}", file=sys.stderr)
        return 0
    text = SERVER.read_text(encoding="utf-8", errors="replace")
    m = re.search(
        r"def _initialize_tool_effectiveness\(self\).*?return\s*\{",
        text,
        re.DOTALL,
    )
    if not m:
        print("could not find _initialize_tool_effectiveness", file=sys.stderr)
        return 1
    start = m.end() - 1
    depth = 0
    end = start
    for i, ch in enumerate(text[start:], start):
        if ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                end = i + 1
                break
    blob = text[start:end]
    # Replace enum refs with string literals for ast parsing
    blob = re.sub(r"TargetType\.([A-Z_]+)\.value", lambda m: json.dumps(TYPE_MAP.get(m.group(1).lower(), m.group(1).lower())), blob)
    blob = blob.replace("TargetType.WEB_APPLICATION.value", '"web"')
    blob = blob.replace("TargetType.NETWORK_HOST.value", '"ip"')
    blob = blob.replace("TargetType.API_ENDPOINT.value", '"api"')
    blob = blob.replace("TargetType.CLOUD_SERVICE.value", '"cloud"')
    blob = blob.replace("TargetType.BINARY_FILE.value", '"binary"')
    # strip comments
    blob = re.sub(r"#.*", "", blob)
    data = ast.literal_eval(blob)
    mapped: dict[str, dict[str, float]] = {}
    for k, v in data.items():
        key = TYPE_MAP.get(str(k).lower(), str(k))
        if key in ("web_application", "network_host", "api_endpoint", "cloud_service", "binary_file"):
            key = TYPE_MAP[key]
        mapped[key] = {str(t): float(s) for t, s in v.items()}
    OUT_JSON.parent.mkdir(parents=True, exist_ok=True)
    OUT_JSON.write_text(json.dumps(mapped, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(f"wrote {OUT_JSON} ({sum(len(v) for v in mapped.values())} scores)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
