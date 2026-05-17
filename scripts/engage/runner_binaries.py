"""Canonical engage-runner PATH binaries — sync with runner.Dockerfile and list-runner-binaries.sh."""

from __future__ import annotations

# Tier-1 + P9g runner-full headless wrappers (generate-tools-live.py / generate-tools-na-matrix.py).
RUNNER_BINARIES_TIER1 = frozenset({
    "nmap", "masscan", "sqlmap", "nikto", "gobuster", "feroxbuster",
    "nuclei", "httpx", "subfinder", "katana", "naabu", "dnsx", "gau",
    "waybackurls", "dalfox", "amass", "ffuf", "arjun", "dirsearch",
    "paramspider", "rustscan", "trivy", "dnsenum", "fierce", "hydra",
    "wafw00f", "enum4linux", "sslscan", "testssl", "dirb",
    "whatweb", "nbtscan", "binwalk", "jaeles", "x8", "enum4linux-ng",
})

# P9g engage-runner-full only (also listed in RUNNER_BINARIES for matrix classification).
RUNNER_BINARIES_FULL = frozenset({
    "burpsuite", "ghidra", "hashcat", "john", "gdb", "metasploit",
    "angr", "radare2", "volatility", "wpscan",
})

# P9i: remaining catalog subprocess binaries (~57).
RUNNER_BINARIES_P9I = frozenset({
    "anew", "arp", "correlate", "delete", "detect", "discover", "display",
    "docker", "dotdotpwn", "error", "exiftool", "falco", "foremost", "format",
    "graphql", "hakrawler", "hashpump", "install", "intelligent", "jwt", "libc",
    "modify", "monitor", "msfvenom", "netexec", "objdump", "one", "optimize",
    "pacu", "pause", "prowler", "pwninit", "pwntools", "qsreplace", "research",
    "responder", "resume", "ropgadget", "ropper", "rpcclient", "scout", "select",
    "server", "smbmap", "steghide", "strings", "terminate", "terrascan", "test",
    "threat", "uro", "volatility3", "vulnerability", "wfuzz", "xsser", "xxd", "zap",
})

# P10b: PythonEnvironmentManager wrappers on engage-runner PATH.
RUNNER_BINARIES_P10B = frozenset({
    "engage-python-install",
    "engage-python-exec",
})

RUNNER_BINARIES = (
    RUNNER_BINARIES_TIER1 | RUNNER_BINARIES_FULL | RUNNER_BINARIES_P9I | RUNNER_BINARIES_P10B
)
