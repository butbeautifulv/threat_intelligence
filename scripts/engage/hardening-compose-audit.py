#!/usr/bin/env python3
"""Static compose/profile hardening audit — no network, no container attacks."""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
FINDINGS: list[tuple[str, str, str]] = []


def add(sev: str, fid: str, msg: str) -> None:
    FINDINGS.append((sev, fid, msg))


def read(p: Path) -> str:
    return p.read_text(encoding="utf-8", errors="replace")


def audit_secure_profiles() -> None:
    for name in ("secure-engage.env", "secure-graph.env"):
        p = ROOT / "deploy/profiles" / name
        if not p.exists():
            add("medium", f"missing-{name}", f"profile not found: {p}")
            continue
        body = read(p)
        if "ENGAGE_DENY_RAW_COMMAND=1" not in body and "secure-engage" in name:
            add("high", "engage-no-deny-raw", f"{name} must set ENGAGE_DENY_RAW_COMMAND=1")
        if "VEIL_REQUIRE_AUTH=1" not in body and "secure-graph" in name:
            add("high", "graph-no-require-auth", f"{name} must set VEIL_REQUIRE_AUTH=1")
        if re.search(r"ENGAGE_ALLOW_RAW_COMMAND\s*=\s*1", body):
            add("high", "profile-allow-raw", f"{name} must not enable ENGAGE_ALLOW_RAW_COMMAND")


def audit_compose_privileged() -> None:
    for yml in (ROOT / "deploy").rglob("compose*.yml"):
        body = read(yml)
        if "privileged: true" in body and "engage" in str(yml):
            add("high", "privileged-engage", f"{yml.relative_to(ROOT)}: privileged: true on engage stack")


def audit_published_neo4j_in_secure() -> None:
    secure = ROOT / "deploy/knowledge/compose.secure.yml"
    if secure.exists():
        body = read(secure)
        if re.search(r"7474|7687", body) and "ports:" in body:
            # secure overlay should not publish neo4j
            if "neo4j" in body.lower() and "ports:" in body.split("neo4j")[-1][:400]:
                add("medium", "secure-neo4j-ports", "compose.secure.yml may publish Neo4j ports")


def main() -> int:
    audit_secure_profiles()
    audit_compose_privileged()
    audit_published_neo4j_in_secure()
    highs = [f for f in FINDINGS if f[0] == "high"]
    for sev, fid, msg in FINDINGS:
        print(f"[{sev}] {fid}: {msg}")
    if highs:
        print(f"\nFAIL: {len(highs)} high-severity finding(s)", file=sys.stderr)
        return 1
    print(f"\nOK hardening compose audit ({len(FINDINGS)} non-blocking findings)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
