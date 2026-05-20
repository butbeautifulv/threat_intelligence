#!/usr/bin/env python3
"""Compare legacy HexStrike Flask routes with Veil engage router (Phase audit)."""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
LEGACY = ROOT / ".external/hexstrike-ai-master/hexstrike_server.py"
HTTPSERVER = ROOT / "engage/serve/internal/transport/httpserver"
PKG_API = ROOT / "pkg/api"
OUT_CSV = ROOT / "docs/engage/engage-route-parity.csv"

ROUTE_ALIASES: dict[str, str] = {
    "DELETE /api/files/delete": "POST /api/files/delete",
    "POST /api/ai/generate_payload": "POST /api/payloads/generate",
    "POST /api/ai/advanced-payload-generation": "POST /api/payloads/generate",
    "GET /api/process/cache-stats": "GET /api/cache/stats",
    "POST /api/process/clear-cache": "POST /api/cache/clear",
    "GET /api/process/pool-stats": "GET /api/processes/dashboard",
    "GET /api/process/performance-dashboard": "GET /api/processes/dashboard",
    "GET /api/process/health-check": "GET /health",
    "POST /api/process/terminate-gracefully/{pid}": "POST /api/processes/terminate/{pid}",
    "GET /api/processes/status/{pid}": "GET /api/processes/status/{pid}",
    "POST /api/processes/terminate/{pid}": "POST /api/processes/terminate/{pid}",
    "POST /api/processes/pause/{pid}": "POST /api/processes/pause/{pid}",
    "POST /api/processes/resume/{pid}": "POST /api/processes/resume/{pid}",
}

NA_OUT_OF_SCOPE: dict[str, str] = {
    "POST /api/python/install": "sandbox risk — no Python REPL in engage",
    "POST /api/python/execute": "sandbox risk — use catalog tools / jobs",
    "POST /api/ai/test_payload": "no LLM payload oracle — N/A",
    "POST /api/process/execute-async": "use POST /api/jobs or async smart-scan",
    "POST /api/process/auto-scaling": "N/A — no autoscale controller in engage",
    "POST /api/process/scale-pool": "N/A — no process pool scaling in engage",
    "GET /api/process/get-task-result/{task_id}": "use GET /api/jobs/{id}",
}


def flask_to_go_path(path: str) -> str:
    path = re.sub(r"<int:(\w+)>", r"{\1}", path)
    path = re.sub(r"<(\w+)>", r"{\1}", path)
    return path


def parse_legacy(path: Path) -> list[tuple[str, str]]:
    text = path.read_text(encoding="utf-8", errors="replace")
    routes: list[tuple[str, str]] = []
    for m in re.finditer(
        r'@app\.route\("([^"]+)"\s*,\s*methods=\[([^\]]+)\]',
        text,
    ):
        route_path = flask_to_go_path(m.group(1))
        methods = re.findall(r'"([A-Z]+)"', m.group(2))
        for method in methods:
            routes.append((method, route_path))
    return routes


def parse_register_health_routes() -> set[str]:
    """Detect GET /health via pkg/api.RegisterHealth (used by engage + knowledge routers)."""
    found: set[str] = set()
    register_go = PKG_API / "register.go"
    if not register_go.is_file():
        return found
    text = register_go.read_text(encoding="utf-8")
    if re.search(r'func\s+RegisterHealth\b', text) and re.search(
        r'mux\.HandleFunc\("GET /health"', text
    ):
        found.add("GET /health")
    return found


def parse_engage(dir_path: Path) -> set[str]:
    text = ""
    for p in sorted(dir_path.glob("*.go")):
        text += p.read_text(encoding="utf-8")
    found: set[str] = set()
    patterns = [
        r'postJSON\(mux,\s*"([A-Z]+) ([^"]+)"',
        r'mux\.HandleFunc\("([A-Z]+) ([^"]+)"',
        r'mux\.Handle\("([A-Z]+) ([^"]+)"',
    ]
    for pat in patterns:
        for m in re.finditer(pat, text):
            found.add(f"{m.group(1)} {m.group(2)}")
    for m in re.finditer(r'wf\("(/api/[^"]+)"', text):
        found.add(f"POST {m.group(1)}")
    if re.search(r'\bapi\.RegisterHealth\s*\(', text):
        found.add("GET /health")
    found |= parse_register_health_routes()
    return found


def normalize_key(method: str, path: str) -> str:
    return f"{method.upper()} {flask_to_go_path(path)}"


def classify(method: str, path: str, engage: set[str]) -> tuple[str, str]:
    key = normalize_key(method, path)

    if path.startswith("/api/tools/") and path != "/api/tools":
        return "na_unified_tool", "POST /api/tools/{name}"

    alias = ROUTE_ALIASES.get(key)
    if alias and alias in engage:
        return "implemented", alias

    if key in engage:
        return "implemented", key

    if key in NA_OUT_OF_SCOPE:
        return "na_out_of_scope", NA_OUT_OF_SCOPE[key]

    if re.match(r"^/api/tools/[a-z0-9_-]+$", path, re.I):
        return "na_unified_tool", "POST /api/tools/{name}"

    return "missing", ""


def main() -> int:
    if not LEGACY.is_file():
        print(f"skip: missing {LEGACY}", file=sys.stderr)
        return 0
    legacy_routes = parse_legacy(LEGACY)
    engage = parse_engage(HTTPSERVER)
    counts: dict[str, int] = {}
    missing: list[str] = []
    rows: list[tuple[str, str, str, str]] = []

    for method, path in sorted(set(legacy_routes)):
        status, note = classify(method, path, engage)
        counts[status] = counts.get(status, 0) + 1
        rows.append((method, path, status, note))
        if status == "missing":
            missing.append(normalize_key(method, path))

    print(f"legacy routes: {len(set(legacy_routes))}")
    print(f"engage routes: {len(engage)}")
    for k in sorted(counts):
        print(f"  {k}: {counts[k]}")

    OUT_CSV.parent.mkdir(parents=True, exist_ok=True)
    with OUT_CSV.open("w", encoding="utf-8") as f:
        f.write("method,path,status,note\n")
        for method, path, status, note in rows:
            note_esc = note.replace('"', '""')
            f.write(f'{method},{path},{status},"{note_esc}"\n')
    print(f"wrote {OUT_CSV}")

    if missing:
        print("unexplained missing routes:", missing, file=sys.stderr)
        return 1
    print("route parity OK (0 unexplained missing)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
