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

# Phase 19: map catalog name prefix → runner binary when installed in engage-runner.
PREFIX_BINARY = [
    ("nmap", "nmap"), ("nuclei", "nuclei"), ("httpx", "httpx"), ("ffuf", "ffuf"),
    ("gobuster", "gobuster"), ("nikto", "nikto"), ("sqlmap", "sqlmap"),
    ("subfinder", "subfinder"), ("amass", "amass"), ("katana", "katana"),
    ("gau", "gau"), ("feroxbuster", "feroxbuster"), ("dalfox", "dalfox"),
    ("arjun", "arjun"), ("dirsearch", "dirsearch"), ("paramspider", "paramspider"),
    ("masscan", "masscan"), ("rustscan", "rustscan"), ("trivy", "trivy"),
    ("waybackurls", "waybackurls"), ("dnsenum", "dnsenum"), ("fierce", "fierce"),
    ("hydra", "hydra"), ("wafw00f", "wafw00f"), ("enum4linux", "enum4linux"),
    ("dirb", "dirb"), ("naabu", "naabu"), ("dnsx", "dnsx"),
]


def resolve_binary(name: str, cat: str) -> str:
    low = name.lower()
    for prefix, binary in sorted(PREFIX_BINARY, key=lambda x: -len(x[0])):
        if low == prefix or low.startswith(prefix + "_"):
            return binary
    binary = name.split("_")[0] if "_" in name else name
    if len(binary) > 20:
        binary = DEFAULT_BINARY.get(cat, "echo")
    return binary

GENERIC_ARGS = ["{target}", "{additional_args}"]

# In-process / non-CLI tools: generic args are intentional.
DOCUMENTED_GENERIC = frozenset({
    "analyze_target_intelligence", "analyze_target", "select_optimal_tools",
    "intelligent_smart_scan", "smart_scan", "create_attack_chain",
    "discover_attack_chains", "correlate_threat_intelligence", "optimize_tool_parameters",
    "create_vulnerability_report", "generate_exploit", "generate_payload",
    "advanced_payload_generation", "ai_vulnerability_assessment", "ai_reconnaissance_workflow",
    "bugbounty_reconnaissance_workflow", "bugbounty_vulnerability_hunting",
    "bugbounty_business_logic_testing", "bugbounty_osint_gathering",
    "bugbounty_file_upload_testing", "bugbounty_comprehensive_assessment",
    "bugbounty_authentication_bypass_testing",
    "browser_agent_inspect", "playwright_navigate", "selenium_navigate",
    "api_schema_analyzer", "api_fuzzer", "graphql_scanner", "jwt_analyzer",
    "burpsuite_scan", "checksec_analyze", "ghidra_analysis", "gdb_analyze",
    "gdb_peda_debug", "radare2_scan", "ropper_scan", "pwntools_exploit",
    "msfvenom_generate", "metasploit_run", "msfconsole_execute",
    "arp_scan_discovery", "autorecon_comprehensive", "autorecon_scan",
    "binwalk_analyze", "clair_vulnerability_scan", "exiftool_extract",
    "nbtscan_netbios", "netexec_scan", "smbmap_scan",
    "volatility_analyze", "volatility3_analyze", "vulnerability_intelligence_dashboard",
    "target_timeline_intelligence",
    "ctf_create_challenge_workflow", "ctf_auto_solve_challenge", "ctf_suggest_tools",
    "ctf_team_strategy", "ctf_cryptography_solver", "ctf_forensics_analyzer", "ctf_binary_analyzer",
})

# Engage-only MCP bridge tools (not in legacy hexstrike_mcp.py).
ENGAGE_BRIDGE_TOOLS: dict[str, list[dict]] = {
    "target_timeline_intelligence": [
        {"name": "target", "type": "string", "required": True},
        {"name": "limit", "type": "string", "default": "50", "required": False},
        {"name": "include_graph", "type": "string", "default": "true", "required": False},
    ],
    "ctf_create_challenge_workflow": [
        {"name": "target", "type": "string", "required": True},
        {"name": "name", "type": "string", "required": False},
        {"name": "category", "type": "string", "required": False},
        {"name": "description", "type": "string", "required": False},
    ],
    "ctf_auto_solve_challenge": [
        {"name": "target", "type": "string", "required": True},
        {"name": "execute_tools", "type": "string", "default": "true", "required": False},
        {"name": "max_steps", "type": "string", "default": "8", "required": False},
    ],
    "ctf_suggest_tools": [
        {"name": "target", "type": "string", "required": True},
        {"name": "description", "type": "string", "required": True},
        {"name": "category", "type": "string", "default": "misc", "required": False},
    ],
    "ctf_team_strategy": [
        {"name": "target", "type": "string", "required": True},
    ],
    "ctf_cryptography_solver": [
        {"name": "cipher_text", "type": "string", "required": True},
        {"name": "cipher_type", "type": "string", "default": "unknown", "required": False},
    ],
    "ctf_forensics_analyzer": [
        {"name": "file_path", "type": "string", "required": True},
    ],
    "ctf_binary_analyzer": [
        {"name": "binary_path", "type": "string", "required": True},
    ],
}

# Category CLI defaults (non-generic vs bare target+args).
CATEGORY_ARGS: dict[str, list[str]] = {
    "network": ["{scan_type}", "-p", "{ports}", "{additional_args}", "{target}"],
    "web": ["-u", "{target}", "{additional_args}"],
    "cloud": ["scan", "{target}", "{additional_args}"],
    "auth": ["-l", "{username}", "{target}", "{additional_args}"],
    "osint": ["-d", "{target}", "{additional_args}"],
    "binary": ["-f", "{target}", "{additional_args}"],
    "ctf": ["--challenge", "{target}", "{additional_args}"],
    "intelligence": GENERIC_ARGS,
}

# Per-tool arg templates when parameters imply structured CLI.
ARGS_TEMPLATES: dict[str, list[str]] = {
    "nmap_scan": ["{scan_type}", "-p", "{ports}", "{additional_args}", "{target}"],
    "nmap_advanced_scan": [
        "{scan_type}", "-p", "{ports}", "--script", "{nse_scripts}",
        "{additional_args}", "{target}",
    ],
    "rustscan_fast_scan": ["-a", "{target}", "-p", "{ports}", "{additional_args}"],
    "masscan_high_speed": ["{target}", "-p", "{ports}", "--rate", "{rate}", "{additional_args}"],
    "nuclei_scan": ["-u", "{target}", "-t", "{templates}", "-duc", "{additional_args}"],
    "nuclei_critical_scan": ["-u", "{target}", "-severity", "critical,high", "-duc", "{additional_args}"],
    "nuclei_web_scan": ["-u", "{target}", "-tags", "cve,tech", "-duc", "{additional_args}"],
    "httpx_probe": ["-u", "{target}", "{additional_args}"],
    "gobuster_scan": ["{mode}", "-u", "{target}", "-w", "{wordlist}", "{additional_args}"],
    "nikto_scan": ["-h", "{target}", "{additional_args}"],
    "sqlmap_scan": ["-u", "{target}", "--batch", "--data", "{data}", "{additional_args}"],
    "sqlmap_get_scan": ["-u", "{target}", "--batch", "{additional_args}"],
    "ffuf_scan": ["-u", "{target}", "-w", "{wordlist}", "{additional_args}"],
    "feroxbuster_scan": ["-u", "{target}", "-w", "{wordlist}", "-t", "{threads}", "{additional_args}"],
    "dirb_scan": ["{target}", "-w", "{wordlist}", "{additional_args}"],
    "wpscan_analyze": ["--url", "{target}", "{additional_args}"],
    "subfinder_scan": ["-d", "{target}", "{additional_args}"],
    "amass_scan": ["{mode}", "-d", "{target}", "{additional_args}"],
    "trivy_scan": ["image", "{target}", "{additional_args}"],
    "hydra_attack": [
        "-l", "{username}", "-P", "{password_file}", "{target}", "-s", "{service}", "{additional_args}",
    ],
    "dirsearch_scan": ["-u", "{target}", "-e", "{extensions}", "{additional_args}"],
    "wfuzz_scan": ["-u", "{target}", "-w", "{wordlist}", "{additional_args}"],
    "fierce_scan": ["--domain", "{target}", "{additional_args}"],
    "john_crack": ["{hash_file}", "--wordlist={wordlist}", "{additional_args}"],
    "hashcat_crack": ["-m", "{hash_type}", "-a", "{attack_mode}", "{hash_file}", "{wordlist}", "{additional_args}"],
    "arjun_scan": ["-u", "{target}", "-m", "{method}", "{additional_args}"],
    "prowler_scan": ["{provider}", "--profile", "{profile}", "{additional_args}"],
    "kube_hunter_scan": ["--remote", "{remote}", "{additional_args}"],
    "dalfox_scan": ["url", "{target}", "{additional_args}"],
    "katana_scan": ["-u", "{target}", "{additional_args}"],
    "paramspider_scan": ["-d", "{target}", "{additional_args}"],
    "paramspider_discovery": ["-d", "{target}", "{additional_args}"],
    "paramspider_mine": ["-d", "{target}", "{additional_args}"],
    "wafw00f_scan": ["-a", "{target}", "{additional_args}"],
    "dnsenum_scan": ["--domain", "{target}", "{additional_args}"],
    "enum4linux_scan": ["-a", "{target}", "{additional_args}"],
    "enum4linux_ng_advanced": ["{target}", "-a", "{additional_args}"],
    "smbmap_scan": ["{target}", "{additional_args}"],
    "netexec_scan": ["{target}", "{additional_args}"],
    "scout_suite_assessment": ["{provider}", "--profile", "{profile}", "{additional_args}"],
    "cloudmapper_analysis": ["{action}", "--account", "{account}", "{additional_args}"],
    "docker_bench_security_scan": ["{additional_args}"],
    "kube_bench_cis": ["{targets}", "{additional_args}"],
    "checkov_scan": ["-d", "{target}", "{additional_args}"],
    "terrascan_scan": ["scan", "-d", "{target}", "{additional_args}"],
    "autorecon_scan": ["{target}", "{additional_args}"],
    "responder_scan": ["-I", "{interface}", "{additional_args}"],
    "volatility_scan": ["-f", "{target}", "{additional_args}"],
    "binwalk_scan": ["{target}", "{additional_args}"],
    "strings_analyze": ["{target}", "{additional_args}"],
    "objdump_analyze": ["-d", "{target}", "{additional_args}"],
    "gdb_analyze": ["{target}", "{additional_args}"],
    "radare2_scan": ["{target}", "{additional_args}"],
    "ghidra_analyze": ["{target}", "{additional_args}"],
    "msfvenom_generate": ["-p", "{payload}", "{additional_args}"],
    "theharvester_scan": ["-d", "{target}", "{additional_args}"],
    "recon_ng_run": ["{additional_args}"],
    "spiderfoot_scan": ["{target}", "{additional_args}"],
    # Extended templates (Phase 9 — catalog execution depth)
    "naabu_port_scan": ["-host", "{target}", "-p", "{ports}", "{additional_args}"],
    "dnsx_resolve": ["-d", "{target}", "{additional_args}"],
    "gau_discovery": ["--subs", "{target}", "{additional_args}"],
    "waybackurls_scan": ["{target}", "{additional_args}"],
    "hakrawler_crawl": ["-url", "{target}", "{additional_args}"],
    "katana_crawl": ["-u", "{target}", "{additional_args}"],
    "xsstrike_scan": ["-u", "{target}", "{additional_args}"],
    "commix_scan": ["-u", "{target}", "{additional_args}"],
    "dalfox_xss_scan": ["url", "{target}", "{additional_args}"],
    "xsser_scan": ["--url", "{target}", "{additional_args}"],
    "medusa_attack": ["-h", "{target}", "-u", "{username}", "-P", "{password_file}", "-M", "{service}", "{additional_args}"],
    "ncrack_scan": ["-p", "{ports}", "{target}", "{additional_args}"],
    "searchsploit_search": ["{query}", "{additional_args}"],
    "sslscan_probe": ["{target}", "{additional_args}"],
    "testssl_scan": ["{target}", "{additional_args}"],
    "snmpwalk_scan": ["-v", "{version}", "-c", "{community}", "{target}", "{additional_args}"],
    "semgrep_scan": ["--config", "{config}", "{target}", "{additional_args}"],
    "bandit_scan": ["-r", "{target}", "{additional_args}"],
    "gitleaks_scan": ["detect", "--source", "{target}", "{additional_args}"],
    "trufflehog_scan": ["filesystem", "{target}", "{additional_args}"],
    "aquatone_scan": ["-scan-timeout", "{timeout}", "{target}", "{additional_args}"],
    "eyewitness_capture": ["--web", "--single", "{target}", "{additional_args}"],
    "autorecon_comprehensive": ["{target}", "{additional_args}"],
    "binwalk_analyze": ["{target}", "{additional_args}"],
    "clair_vulnerability_scan": ["{target}", "{additional_args}"],
    "checkov_iac_scan": ["-d", "{target}", "{additional_args}"],
    "ghidra_analysis": ["{target}", "{additional_args}"],
    "gdb_peda_debug": ["{target}", "{additional_args}"],
    "foremost_carving": ["-i", "{target}", "{additional_args}"],
    "exiftool_extract": ["{target}", "{additional_args}"],
    "dotdotpwn_scan": ["-m", "{mode}", "-h", "{target}", "{additional_args}"],
    "api_fuzzer": ["-u", "{target}", "{additional_args}"],
    "burpsuite_scan": ["{target}", "{additional_args}"],
    "arp_scan_discovery": ["{target}", "{additional_args}"],
    "nbtscan_netbios": ["{target}", "{additional_args}"],
    "rpcclient_enum": ["{target}", "{additional_args}"],
    "impacket_secretsdump": ["{target}", "{additional_args}"],
    "crackmapexec_scan": ["{target}", "{additional_args}"],
    "bloodhound_ingest": ["-c", "{collection}", "{additional_args}"],
    "pacu_exploit": ["{module}", "{additional_args}"],
    "cloudsploit_scan": ["{provider}", "{additional_args}"],
    "falco_runtime_monitoring": ["{additional_args}"],
    "kubeaudit_scan": ["{target}", "{additional_args}"],
    "lynis_audit": ["{additional_args}"],
    "openvas_scan": ["{target}", "{additional_args}"],
    "zap_scan": ["-cmd", "-quickurl", "{target}", "{additional_args}"],
    "wpscan_enum": ["--url", "{target}", "{additional_args}"],
    "cewl_wordlist": ["-u", "{target}", "-o", "{output}", "{additional_args}"],
    "crunch_generate": ["{min}", "{max}", "{charset}", "-o", "{output}", "{additional_args}"],
    "hashid_identify": ["{hash}", "{additional_args}"],
    "ophcrack_crack": ["{target}", "{additional_args}"],
    "aircrack_crack": ["{capture_file}", "{additional_args}"],
    "bettercap_attack": ["-eval", "{script}", "{additional_args}"],
    "mitmproxy_intercept": ["{target}", "{additional_args}"],
    "responder_poison": ["-I", "{interface}", "{additional_args}"],
    "enum4linux_scan": ["-a", "{target}", "{additional_args}"],
    "sherlock_username": ["{username}", "{additional_args}"],
    "maigret_username": ["{username}", "{additional_args}"],
    "holehe_email": ["{email}", "{additional_args}"],
    "social_mapper": ["{target}", "{additional_args}"],
    "shodan_search": ["{query}", "{additional_args}"],
    "censys_search": ["{query}", "{additional_args}"],
    "metasploit_run": ["-r", "{resource}", "{additional_args}"],
    "msfconsole_execute": ["-r", "{resource}", "{additional_args}"],
    "nuclei_workflow": ["-u", "{target}", "-w", "{workflow}", "{additional_args}"],
    "httprobe_alive": ["{target}", "{additional_args}"],
    "tlsx_scan": ["-u", "{target}", "{additional_args}"],
    "cdncheck_scan": ["-i", "{target}", "{additional_args}"],
    "asnmap_lookup": ["-d", "{target}", "{additional_args}"],
    "mapcidr_expand": ["-cidr", "{target}", "{additional_args}"],
    "uncover_search": ["-q", "{query}", "{additional_args}"],
    "notify_send": ["-data", "{target}", "{additional_args}"],
    "interactsh_oob": ["{additional_args}"],
    "browser_agent_inspect": ["{target}", "{additional_args}"],
    "playwright_navigate": ["{target}", "{additional_args}"],
    "selenium_navigate": ["{target}", "{additional_args}"],
}

# Tools that may use infer_args_template when not in ARGS_TEMPLATES.
INFER_TOOLS = frozenset({
    "nmap_scan", "nmap_advanced_scan", "rustscan_fast_scan", "masscan_high_speed",
    "nuclei_scan", "httpx_probe", "gobuster_scan", "nikto_scan", "sqlmap_scan",
    "ffuf_scan", "feroxbuster_scan", "dirb_scan", "wpscan_analyze",
    "subfinder_scan", "amass_scan", "trivy_scan", "hydra_attack",
    "prowler_scan", "kube_hunter_scan", "john_crack", "hashcat_crack",
    "dirsearch_scan", "wfuzz_scan", "fierce_scan", "arjun_scan",
    "naabu_port_scan", "dnsenum_scan", "gau_discovery", "xsstrike_scan",
    "medusa_attack", "sslscan_probe", "testssl_scan", "semgrep_scan",
    "enum4linux_ng_advanced", "autorecon_comprehensive", "browser_agent_inspect",
})


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


def infer_args_template(name: str, params: list[dict]) -> list[str]:
    """Heuristic CLI templates from MCP parameter names (allowlist only)."""
    names = {p["name"] for p in params}
    if "scan_type" in names and "ports" in names:
        if "nse_scripts" in names:
            return ARGS_TEMPLATES.get("nmap_advanced_scan", [
                "{scan_type}", "-p", "{ports}", "--script", "{nse_scripts}",
                "{additional_args}", "{target}",
            ])
        return ARGS_TEMPLATES.get("nmap_scan", [
            "{scan_type}", "-p", "{ports}", "{additional_args}", "{target}",
        ])
    if "rate" in names and "ports" in names:
        return ["{target}", "-p", "{ports}", "--rate", "{rate}", "{additional_args}"]
    if "service" in names and "username" in names:
        return [
            "-l", "{username}", "-P", "{password_file}", "{target}", "-s", "{service}",
            "{additional_args}",
        ]
    if "wordlist" in names and ("url" in names or "target" in names):
        if "threads" in names:
            return ["-u", "{target}", "-w", "{wordlist}", "-t", "{threads}", "{additional_args}"]
        if "mode" in names and "match_codes" in names:
            return ["-u", "{target}", "-w", "{wordlist}", "{additional_args}"]
        return ["{target}", "-w", "{wordlist}", "{additional_args}"]
    if "templates" in names:
        return ["-u", "{target}", "-t", "{templates}", "{additional_args}"]
    if "domain" in names and "mode" in names:
        return ["{mode}", "-d", "{target}", "{additional_args}"]
    if "hash_file" in names and "wordlist" in names:
        return ["{hash_file}", "--wordlist={wordlist}", "{additional_args}"]
    if name.endswith("_scan") and "target" in names:
        return ["-h", "{target}", "{additional_args}"] if "nikto" in name else GENERIC_ARGS
    return GENERIC_ARGS


def resolve_args_template(name: str, params: list[dict]) -> list[str]:
    if name in ARGS_TEMPLATES:
        return ARGS_TEMPLATES[name]
    if name in INFER_TOOLS:
        tpl = infer_args_template(name, params)
        if tpl != GENERIC_ARGS:
            return tpl
        # Fall through to category defaults when infer yields generic.
    if name in DOCUMENTED_GENERIC:
        return GENERIC_ARGS
    cat = category_for(name)
    return CATEGORY_ARGS.get(cat, GENERIC_ARGS)


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
    for name, params in ENGAGE_BRIDGE_TOOLS.items():
        param_map.setdefault(name, params)
    names = sorted(param_map.keys())

    lines = [
        "# Veil engage tool catalog (names aligned with legacy MCP reference in .external/)",
        "# Regenerate: make catalog-engage",
        "# enabled=false until runner image provides the binary on PATH.",
        "tools:",
    ]
    non_generic = 0
    for name in names:
        cat = category_for(name)
        if name in ENGAGE_BRIDGE_TOOLS:
            cat = "ctf" if name.startswith("ctf_") else "intelligence"
            binary = "api"
        else:
            binary = resolve_binary(name, cat)
        params = param_map.get(name, [{"name": "target", "type": "string", "required": True}])
        args = resolve_args_template(name, params)
        if args != GENERIC_ARGS:
            non_generic += 1

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
        lines.append("    timeout_sec: 300")
        lines.append(f"    description: {yaml_quote(describe(name, cat))}")
        lines.append("    enabled: false")

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"wrote {len(names)} tools to {OUT} ({non_generic} non-generic args templates)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
