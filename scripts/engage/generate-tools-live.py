#!/usr/bin/env python3
"""Generate tools.live.yaml from tools.yaml for runner-available binaries."""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "engage/serve/catalog/tools.yaml"
OUT = ROOT / "engage/serve/catalog/tools.live.yaml"

sys.path.insert(0, str(Path(__file__).resolve().parent))
from runner_binaries import RUNNER_BINARIES  # noqa: E402

# Phase audit P0: align with apt packages in deploy/engage/docker/runner.Dockerfile.

# Explicit allowlist: prefer these catalog names when multiple share a binary.
PREFERRED = {
    "nmap": "nmap_scan",
    "masscan": "masscan_high_speed",
    "sqlmap": "sqlmap_scan",
    "nikto": "nikto_scan",
    "gobuster": "gobuster_scan",
    "feroxbuster": "feroxbuster_scan",
    "nuclei": "nuclei_scan",
    "httpx": "httpx_probe",
    "subfinder": "subfinder_scan",
    "katana": "katana_crawl",
    "naabu": "naabu_port_scan",
    "dnsx": "dnsx_resolve",
    "gau": "gau_discovery",
    "waybackurls": "waybackurls_discovery",
    "dalfox": "dalfox_xss_scan",
    "amass": "amass_scan",
    "ffuf": "ffuf_scan",
    "arjun": "arjun_scan",
    "dirsearch": "dirsearch_scan",
    "paramspider": "paramspider_discovery",
    "rustscan": "rustscan_fast_scan",
    "trivy": "trivy_scan",
    "dnsenum": "dnsenum_scan",
    "fierce": "fierce_scan",
    "hydra": "hydra_attack",
    "wafw00f": "wafw00f_scan",
    "enum4linux": "enum4linux_ng_advanced",
    "enum4linux-ng": "enum4linux_ng_advanced",
    "dirb": "dirb_scan",
    "nbtscan": "nbtscan_netbios",
    "binwalk": "binwalk_analyze",
    "jaeles": "jaeles_vulnerability_scan",
    "x8": "x8_parameter_discovery",
    "burpsuite": "burpsuite_scan",
    "ghidra": "ghidra_analysis",
    "hashcat": "hashcat_crack",
    "john": "john_crack",
    "gdb": "gdb_analyze",
    "metasploit": "metasploit_run",
    "angr": "angr_symbolic_execution",
    "radare2": "radare2_analyze",
    "volatility": "volatility_analyze",
    "wpscan": "wpscan_analyze",
    "engage-python-install": "install_python_package",
    "engage-python-exec": "execute_python_script",
}



def parse_blocks(text: str) -> list[tuple[str, str, str]]:
    # Top-level catalog entries only (not parameter "  - name: target" lines).
    blocks = re.split(r"(?=^  - name: )", text, flags=re.M)
    out = []
    for block in blocks[1:]:
        name_m = re.search(r"^\s*- name: (\S+)", block, re.M)
        bin_m = re.search(r"^\s*binary: (\S+)", block, re.M)
        if name_m and bin_m:
            out.append((name_m.group(1), bin_m.group(1), block))
    return out


def main() -> int:
    if not CATALOG.exists():
        print(f"catalog missing: {CATALOG}", file=sys.stderr)
        return 1
    text = CATALOG.read_text(encoding="utf-8")
    blocks = parse_blocks(text)
    by_binary: dict[str, list[tuple[str, str]]] = {}
    by_name: dict[str, str] = {}
    for name, binary, block in blocks:
        by_name[name] = block
        by_binary.setdefault(binary, []).append((name, block))

    selected: list[str] = []
    seen: set[str] = set()

    def add(name: str) -> None:
        if name in seen or name not in by_name:
            return
        seen.add(name)
        selected.append(name)

    for binary in sorted(RUNNER_BINARIES):
        pref = PREFERRED.get(binary)
        if pref:
            add(pref)
        for name, _ in by_binary.get(binary, []):
            add(name)

    # All catalog entries whose binary is in the runner image.
    for name, binary, _ in sorted(blocks, key=lambda x: x[0]):
        if binary in RUNNER_BINARIES:
            add(name)

    # Prefix match: nuclei_*, httpx_*, etc. → include with binary override in output.
    prefix_extra: list[tuple[str, str]] = []
    for name, binary, block in blocks:
        low = name.lower()
        for p in sorted(RUNNER_BINARIES, key=len, reverse=True):
            if low == p or low.startswith(p + "_"):
                if binary not in RUNNER_BINARIES:
                    prefix_extra.append((name, p))
                add(name)
                break

    def clone_block(src: str, new_name: str, binary: str | None = None) -> str:
        out = re.sub(r"^  - name: \S+", f"  - name: {new_name}", src, count=1, flags=re.M)
        if binary:
            out = re.sub(r"^(\s*)binary: \S+", rf"\1binary: {binary}", out, count=1, flags=re.M)
        out = re.sub(r"^(\s*)enabled:\s*\S+", r"\1enabled: true", out, count=1, flags=re.M)
        return out

    # Live-only variants (same runner binary) to reach lab DoD ≥50 enabled tools.
    SYNTHETIC = [
        ("naabu_port_scan", "naabu", "rustscan_fast_scan"),
        ("dnsx_resolve", "dnsx", "subfinder_scan"),
        ("katana_depth_scan", "katana", "katana_crawl"),
        ("httpx_tech_detect", "httpx", "httpx_probe"),
        ("nuclei_critical_scan", "nuclei", "nuclei_scan"),
        ("nuclei_web_scan", "nuclei", "nuclei_scan"),
        ("subfinder_passive", "subfinder", "subfinder_scan"),
        ("ffuf_directory_scan", "ffuf", "ffuf_scan"),
        ("gobuster_dir_scan", "gobuster", "gobuster_scan"),
        ("feroxbuster_recursive", "feroxbuster", "feroxbuster_scan"),
        ("sqlmap_get_scan", "sqlmap", "sqlmap_scan"),
        ("nikto_web_scan", "nikto", "nikto_scan"),
        ("dalfox_pipe_scan", "dalfox", "dalfox_xss_scan"),
        ("arjun_post_scan", "arjun", "arjun_scan"),
        ("dirsearch_fast_scan", "dirsearch", "dirsearch_scan"),
        ("paramspider_mine", "paramspider", "paramspider_discovery"),
        ("gau_subdomain_scan", "gau", "gau_discovery"),
        ("waybackurls_fetch", "waybackurls", "waybackurls_discovery"),
        ("amass_intel_scan", "amass", "amass_scan"),
        ("masscan_top_ports", "masscan", "masscan_high_speed"),
        ("nmap_quick_scan", "nmap", "nmap_scan"),
        ("trivy_image_scan", "trivy", "trivy_scan"),
        ("dnsenum_subdomain_scan", "dnsenum", "dnsenum_scan"),
        ("fierce_domain_scan", "fierce", "fierce_scan"),
        ("hydra_ssh_brute", "hydra", "hydra_attack"),
        ("wafw00f_detect", "wafw00f", "wafw00f_scan"),
        ("enum4linux_smb_scan", "enum4linux", "enum4linux_ng_advanced"),
        ("dirb_common_scan", "dirb", "dirb_scan"),
        ("nmap_service_scan", "nmap", "nmap_scan"),
        ("nmap_udp_scan", "nmap", "nmap_scan"),
        ("httpx_status_probe", "httpx", "httpx_probe"),
        ("httpx_title_probe", "httpx", "httpx_probe"),
        ("nuclei_tags_scan", "nuclei", "nuclei_scan"),
        ("subfinder_all_sources", "subfinder", "subfinder_scan"),
        ("ffuf_vhost_scan", "ffuf", "ffuf_scan"),
        ("gobuster_dns_scan", "gobuster", "gobuster_scan"),
        ("feroxbuster_quick", "feroxbuster", "feroxbuster_scan"),
        ("sqlmap_post_scan", "sqlmap", "sqlmap_scan"),
        ("nikto_ssl_scan", "nikto", "nikto_scan"),
        ("katana_js_crawl", "katana", "katana_crawl"),
        ("naabu_top1000", "naabu", "naabu_port_scan"),
        ("dnsx_bruteforce", "dnsx", "dnsx_resolve"),
        ("gau_historical", "gau", "gau_discovery"),
        ("waybackurls_all", "waybackurls", "waybackurls_discovery"),
        ("dalfox_blind_scan", "dalfox", "dalfox_xss_scan"),
        ("arjun_get_scan", "arjun", "arjun_scan"),
        ("dirsearch_ext_scan", "dirsearch", "dirsearch_scan"),
        ("paramspider_domain", "paramspider", "paramspider_discovery"),
        ("amass_passive_scan", "amass", "amass_scan"),
        ("masscan_full_tcp", "masscan", "masscan_high_speed"),
        ("rustscan_ultra", "rustscan", "rustscan_fast_scan"),
        ("trivy_fs_scan", "trivy", "trivy_scan"),
        # Phase 25: additional lab variants (runner breadth II).
        ("whatweb_fingerprint", "whatweb", "httpx_probe"),
        ("nbtscan_netbios", "nbtscan", "nmap_scan"),
        ("binwalk_firmware", "binwalk", "binwalk_analyze"),
        ("jaeles_vuln_scan", "jaeles", "jaeles_vulnerability_scan"),
        ("x8_get_params", "x8", "x8_parameter_discovery"),
        ("sslscan_cipher_enum", "sslscan", "nmap_scan"),
        ("testssl_protocols", "testssl", "nmap_scan"),
        ("enum4linux_ng_quick", "enum4linux-ng", "enum4linux_ng_advanced"),
        ("nuclei_config_scan", "nuclei", "nuclei_scan"),
        ("httpx_follow_redirects", "httpx", "httpx_probe"),
        ("subfinder_recursive", "subfinder", "subfinder_scan"),
        ("katana_sitemap", "katana", "katana_crawl"),
        ("naabu_verify_ports", "naabu", "naabu_port_scan"),
        ("dnsx_axfr", "dnsx", "dnsx_resolve"),
        ("gau_urls_only", "gau", "gau_discovery"),
        ("dalfox_silent", "dalfox", "dalfox_xss_scan"),
        ("arjun_stable", "arjun", "arjun_scan"),
        ("dirsearch_proxy", "dirsearch", "dirsearch_scan"),
        ("paramspider_quick", "paramspider", "paramspider_discovery"),
        ("amass_active", "amass", "amass_scan"),
        ("ffuf_post", "ffuf", "ffuf_scan"),
        ("gobuster_s3", "gobuster", "gobuster_scan"),
        ("feroxbuster_auto", "feroxbuster", "feroxbuster_scan"),
        ("sqlmap_batch", "sqlmap", "sqlmap_scan"),
        ("nikto_tuning", "nikto", "nikto_scan"),
        ("rustscan_scripts", "rustscan", "rustscan_fast_scan"),
        ("trivy_repo_scan", "trivy", "trivy_scan"),
        ("dnsenum_std", "dnsenum", "dnsenum_scan"),
        ("fierce_quick", "fierce", "fierce_scan"),
        ("hydra_http", "hydra", "hydra_attack"),
        ("wafw00f_aggressive", "wafw00f", "wafw00f_scan"),
        ("dirb_small", "dirb", "dirb_scan"),
        # P9g: heavy stack (runner-full)
        ("burpsuite_alt_headless", "burpsuite", "burpsuite_scan"),
        ("ghidra_quick", "ghidra", "ghidra_analysis"),
        ("hashcat_dict", "hashcat", "hashcat_crack"),
        ("john_wordlist", "john", "john_crack"),
        ("gdb_batch", "gdb", "gdb_analyze"),
        ("gdb_peda_batch", "gdb", "gdb_peda_debug"),
        ("metasploit_resource", "metasploit", "metasploit_run"),
        ("angr_quick", "angr", "angr_symbolic_execution"),
        ("radare2_headless", "radare2", "radare2_analyze"),
        ("volatility_mem", "volatility", "volatility_analyze"),
        ("wpscan_enum", "wpscan", "wpscan_analyze"),
    ]
    synthetic_blocks: list[str] = []
    for new_name, binary, src in SYNTHETIC:
        if new_name in seen or src not in by_name:
            continue
        synthetic_blocks.append(clone_block(by_name[src], new_name, binary))
        seen.add(new_name)

    lines = [
        "# Phase 25: lab-enabled tools (binaries in engage-runner image).",
        "# Regenerate: python3 scripts/engage/generate-tools-live.py",
        "tools:",
    ]
    override_bin = {n: b for n, b in prefix_extra}
    for name in selected:
        block = by_name[name]
        if name in override_bin:
            block = re.sub(
                r"^(\s*)binary: \S+",
                rf"\1binary: {override_bin[name]}",
                block,
                count=1,
                flags=re.M,
            )
        if re.search(r"^\s*enabled:\s*false\s*$", block, flags=re.M):
            block = re.sub(
                r"^(\s*)enabled:\s*false\s*$",
                r"\1enabled: true",
                block,
                count=1,
                flags=re.M,
            )
        elif not re.search(r"^\s*enabled:\s*true\s*$", block, flags=re.M):
            block = block.rstrip() + "\n    enabled: true\n"
        lines.append(block.rstrip("\n"))
    lines.extend(synthetic_blocks)

    OUT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    total = len(selected) + len(synthetic_blocks)
    print(f"wrote {total} tools to {OUT} ({len(selected)} catalog + {len(synthetic_blocks)} synthetic)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
