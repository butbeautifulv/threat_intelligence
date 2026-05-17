package runner

// CatalogBinaries lists subprocess tool binaries expected on engage-runner PATH.
// Keep in sync with scripts/engage/runner_binaries.py and list-runner-binaries.sh.
var CatalogBinaries = map[string]struct{}{
	"nmap": {}, "masscan": {}, "sqlmap": {}, "nikto": {}, "gobuster": {}, "feroxbuster": {},
	"nuclei": {}, "httpx": {}, "subfinder": {}, "katana": {}, "naabu": {}, "dnsx": {},
	"gau": {}, "waybackurls": {}, "dalfox": {}, "amass": {}, "ffuf": {}, "arjun": {},
	"dirsearch": {}, "paramspider": {}, "rustscan": {}, "trivy": {}, "dnsenum": {},
	"fierce": {}, "hydra": {}, "wafw00f": {}, "enum4linux": {}, "sslscan": {},
	"testssl": {}, "dirb": {}, "whatweb": {}, "nbtscan": {}, "binwalk": {},
	"jaeles": {}, "x8": {}, "enum4linux-ng": {},
	"burpsuite": {}, "ghidra": {}, "hashcat": {}, "john": {}, "gdb": {}, "metasploit": {},
	"angr": {}, "radare2": {}, "volatility": {}, "wpscan": {},
	"anew": {}, "arp": {}, "correlate": {}, "delete": {}, "detect": {}, "discover": {},
	"display": {}, "docker": {}, "dotdotpwn": {}, "error": {}, "exiftool": {}, "falco": {},
	"foremost": {}, "format": {}, "graphql": {}, "hakrawler": {}, "hashpump": {},
	"install": {}, "intelligent": {}, "jwt": {}, "libc": {}, "modify": {}, "monitor": {},
	"msfvenom": {}, "netexec": {}, "objdump": {}, "one": {}, "optimize": {}, "pacu": {},
	"pause": {}, "prowler": {}, "pwninit": {}, "pwntools": {}, "qsreplace": {}, "research": {},
	"responder": {}, "resume": {}, "ropgadget": {}, "ropper": {}, "rpcclient": {}, "scout": {},
	"select": {}, "server": {}, "smbmap": {}, "steghide": {}, "strings": {}, "terminate": {},
	"terrascan": {}, "test": {}, "threat": {}, "uro": {}, "volatility3": {},
	"vulnerability": {}, "wfuzz": {}, "xsser": {}, "xxd": {}, "zap": {},
	"engage-python-install": {}, "engage-python-exec": {},
}

// IsCatalogBinary reports whether name is a known engage-runner catalog binary.
func IsCatalogBinary(name string) bool {
	_, ok := CatalogBinaries[name]
	return ok
}
