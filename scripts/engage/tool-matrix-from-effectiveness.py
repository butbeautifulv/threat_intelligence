#!/usr/bin/env python3
"""Emit tool-matrix.targets from effectiveness_legacy.json (score >= threshold)."""
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
EFFECTIVENESS = (
    ROOT / "pkg/decision/testdata/effectiveness_legacy.json"
)
OUT = ROOT / "scripts/engage/tool-matrix.targets"
THRESHOLD = float(sys.argv[1]) if len(sys.argv) > 1 else 0.85

# binary -> catalog name (runner tier-1)
BINARY_TO_CATALOG = {
    "nmap": "nmap_scan",
    "nuclei": "nuclei_scan",
    "httpx": "httpx_probe",
    "subfinder": "subfinder_scan",
    "gobuster": "gobuster_scan",
    "nikto": "nikto_scan",
    "rustscan": "rustscan_fast_scan",
    "feroxbuster": "feroxbuster_scan",
    "ffuf": "ffuf_scan",
    "sqlmap": "sqlmap_scan",
    "amass": "amass_scan",
    "katana": "katana_crawl",
    "gau": "gau_discovery",
    "arjun": "arjun_scan",
    "paramspider": "paramspider_discovery",
    "dalfox": "dalfox_xss_scan",
    "masscan": "masscan_high_speed",
    "trivy": "trivy_scan",
    "hydra": "hydra_attack",
    "wpscan": "wpscan_analyze",
    "dirsearch": "dirsearch_scan",
    "waybackurls": "waybackurls_discovery",
    "fierce": "fierce_scan",
    "dnsenum": "dnsenum_scan",
    "wafw00f": "wafw00f_scan",
    "enum4linux": "enum4linux_scan",
    "enum4linux-ng": "enum4linux_ng_advanced",
}

SAFE_TARGET = {
    "nmap_scan": "127.0.0.1",
    "nuclei_scan": "https://example.com",
    "httpx_probe": "https://example.com",
    "subfinder_scan": "example.com",
    "gobuster_scan": "http://127.0.0.1",
    "nikto_scan": "127.0.0.1",
    "rustscan_fast_scan": "127.0.0.1",
    "feroxbuster_scan": "http://127.0.0.1",
    "ffuf_scan": "http://127.0.0.1",
    "sqlmap_scan": "http://127.0.0.1",
    "trivy_scan": "alpine:latest",
    "arjun_scan": "https://example.com",
    "dalfox_xss_scan": "https://example.com",
    "masscan_high_speed": "127.0.0.1",
    "amass_scan": "example.com",
    "katana_crawl": "https://example.com",
    "gau_discovery": "example.com",
    "paramspider_discovery": "example.com",
    "dirsearch_scan": "https://example.com",
    "waybackurls_discovery": "example.com",
    "hydra_attack": "127.0.0.1",
    "wpscan_analyze": "https://example.com",
    "fierce_scan": "example.com",
    "dnsenum_scan": "example.com",
    "wafw00f_scan": "https://example.com",
    "enum4linux_scan": "127.0.0.1",
    "enum4linux_ng_advanced": "127.0.0.1",
}


def main() -> int:
    if not EFFECTIVENESS.is_file():
        print(f"missing {EFFECTIVENESS}", file=sys.stderr)
        return 1
    data = json.loads(EFFECTIVENESS.read_text(encoding="utf-8"))
    lines: list[str] = []
    seen: set[str] = set()
    for _ttype, tools in data.items():
        for binary, score in tools.items():
            if score < THRESHOLD:
                continue
            catalog = BINARY_TO_CATALOG.get(binary)
            if not catalog or catalog in seen:
                continue
            target = SAFE_TARGET.get(catalog, "example.com")
            lines.append(f"{catalog}:{target}")
            seen.add(catalog)
    OUT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"wrote {len(lines)} matrix targets to {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
