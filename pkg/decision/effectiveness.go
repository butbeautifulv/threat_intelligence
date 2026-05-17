package decision

// defaultEffectivenessTables mirrors HexStrike IntelligentDecisionEngine tool effectiveness.
// Regenerate JSON snapshot: scripts/engage/extract-decision-tables.py
func defaultEffectivenessTables() map[string]map[string]float64 {
	return map[string]map[string]float64{
		"web": {
			"nmap": 0.8, "gobuster": 0.9, "nuclei": 0.95, "nikto": 0.85, "sqlmap": 0.9,
			"ffuf": 0.9, "feroxbuster": 0.85, "katana": 0.88, "httpx": 0.85, "wpscan": 0.95,
			"burpsuite": 0.9, "dirsearch": 0.87, "gau": 0.82, "waybackurls": 0.8, "arjun": 0.9,
			"paramspider": 0.85, "x8": 0.88, "jaeles": 0.92, "dalfox": 0.93,
			"anew": 0.7, "qsreplace": 0.75, "uro": 0.7,
		},
		"ip": {
			"nmap": 0.95, "nmap-advanced": 0.97, "masscan": 0.92, "rustscan": 0.9,
			"autorecon": 0.95, "enum4linux": 0.8, "enum4linux-ng": 0.88, "smbmap": 0.85,
			"rpcclient": 0.82, "nbtscan": 0.75, "arp-scan": 0.85, "responder": 0.88,
			"hydra": 0.8, "netexec": 0.85, "amass": 0.7,
		},
		"api": {
			"nuclei": 0.9, "ffuf": 0.85, "arjun": 0.95, "paramspider": 0.88,
			"httpx": 0.9, "x8": 0.92, "katana": 0.85, "jaeles": 0.88, "postman": 0.8,
		},
		"cloud": {
			"prowler": 0.95, "scout-suite": 0.92, "cloudmapper": 0.88, "pacu": 0.85,
			"trivy": 0.9, "clair": 0.85, "kube-hunter": 0.9, "kube-bench": 0.88,
			"docker-bench-security": 0.85, "falco": 0.87, "checkov": 0.9, "terrascan": 0.88,
		},
		"binary": {
			"ghidra": 0.95, "radare2": 0.9, "gdb": 0.85, "gdb-peda": 0.92,
			"angr": 0.88, "pwntools": 0.9, "ropgadget": 0.85, "ropper": 0.88,
			"one-gadget": 0.82, "libc-database": 0.8, "checksec": 0.75,
			"strings": 0.7, "objdump": 0.75, "binwalk": 0.8, "pwninit": 0.85,
		},
		"unknown": {
			"nmap": 0.7, "httpx": 0.7, "subfinder": 0.75, "nuclei": 0.8,
		},
	}
}
