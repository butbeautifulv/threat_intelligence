package decision

import "strings"

// AttackStep is one ordered tool invocation in a pattern (HexStrike attack_patterns).
type AttackStep struct {
	Tool     string
	Priority int
	Params   map[string]string
}

// AttackPatterns returns named multi-step scenarios (HexStrike _initialize_attack_patterns parity).
func AttackPatterns() map[string][]AttackStep {
	return map[string][]AttackStep{
		"web_reconnaissance": {
			{Tool: "nmap", Priority: 1, Params: map[string]string{"scan_type": "-sV -sC", "ports": "80,443,8080,8443"}},
			{Tool: "httpx", Priority: 2, Params: map[string]string{"additional_args": "-silent -tech-detect"}},
			{Tool: "katana", Priority: 3, Params: map[string]string{"additional_args": "-silent"}},
			{Tool: "gau", Priority: 4, Params: map[string]string{"additional_args": "--threads 5"}},
			{Tool: "gobuster", Priority: 5, Params: map[string]string{"mode": "dir", "wordlist": "/usr/share/wordlists/dirb/common.txt"}},
			{Tool: "nuclei", Priority: 6, Params: map[string]string{"templates": "cves/,misconfiguration/"}},
			{Tool: "nikto", Priority: 7, Params: map[string]string{"additional_args": "-Tuning 123bde"}},
		},
		"api_testing": {
			{Tool: "httpx", Priority: 1, Params: map[string]string{"additional_args": "-silent"}},
			{Tool: "arjun", Priority: 2, Params: map[string]string{"method": "GET,POST"}},
			{Tool: "paramspider", Priority: 3, Params: map[string]string{"level": "2"}},
			{Tool: "nuclei", Priority: 4, Params: map[string]string{"templates": "exposures/,misconfiguration/"}},
			{Tool: "ffuf", Priority: 5, Params: map[string]string{"wordlist": "/usr/share/wordlists/dirb/common.txt"}},
		},
		"network_discovery": {
			{Tool: "rustscan", Priority: 1, Params: map[string]string{"ports": "1-65535"}},
			{Tool: "nmap", Priority: 2, Params: map[string]string{"scan_type": "-sV"}},
			{Tool: "masscan", Priority: 3, Params: map[string]string{"rate": "1000", "ports": "1-1024"}},
			{Tool: "enum4linux", Priority: 4, Params: nil},
			{Tool: "smbmap", Priority: 5, Params: nil},
		},
		"vulnerability_assessment": {
			{Tool: "nuclei", Priority: 1, Params: map[string]string{"severity": "critical,high,medium"}},
			{Tool: "dalfox", Priority: 2, Params: nil},
			{Tool: "nikto", Priority: 3, Params: nil},
			{Tool: "sqlmap", Priority: 4, Params: map[string]string{"additional_args": "--batch"}},
		},
		"comprehensive_network_pentest": {
			{Tool: "autorecon", Priority: 1, Params: nil},
			{Tool: "rustscan", Priority: 2, Params: nil},
			{Tool: "nmap", Priority: 3, Params: map[string]string{"scan_type": "-sV -sC"}},
			{Tool: "enum4linux", Priority: 4, Params: nil},
			{Tool: "responder", Priority: 5, Params: nil},
		},
		"binary_exploitation": {
			{Tool: "checksec", Priority: 1, Params: nil},
			{Tool: "ghidra", Priority: 2, Params: nil},
			{Tool: "radare2", Priority: 3, Params: nil},
			{Tool: "gdb", Priority: 4, Params: nil},
		},
		"ctf_pwn_challenge": {
			{Tool: "checksec", Priority: 1, Params: nil},
			{Tool: "ghidra", Priority: 2, Params: nil},
			{Tool: "radare2", Priority: 3, Params: nil},
			{Tool: "strings", Priority: 4, Params: nil},
		},
		"ctf_web_challenge": {
			{Tool: "httpx", Priority: 1, Params: map[string]string{"additional_args": "-silent -tech-detect"}},
			{Tool: "katana", Priority: 2, Params: map[string]string{"additional_args": "-silent"}},
			{Tool: "nuclei", Priority: 3, Params: map[string]string{"severity": "critical,high"}},
			{Tool: "gobuster", Priority: 4, Params: nil},
			{Tool: "arjun", Priority: 5, Params: nil},
		},
		"ctf_crypto_challenge": {
			{Tool: "hashcat", Priority: 1, Params: nil},
			{Tool: "john", Priority: 2, Params: nil},
		},
		"ctf_forensics_challenge": {
			{Tool: "binwalk", Priority: 1, Params: nil},
			{Tool: "exiftool", Priority: 2, Params: nil},
			{Tool: "strings", Priority: 3, Params: nil},
		},
		"aws_security_assessment": {
			{Tool: "prowler", Priority: 1, Params: map[string]string{"provider": "aws"}},
			{Tool: "scout-suite", Priority: 2, Params: nil},
			{Tool: "cloudmapper", Priority: 3, Params: nil},
		},
		"kubernetes_security_assessment": {
			{Tool: "kube-bench", Priority: 1, Params: nil},
			{Tool: "kube-hunter", Priority: 2, Params: nil},
			{Tool: "trivy", Priority: 3, Params: nil},
		},
		"container_security_assessment": {
			{Tool: "trivy", Priority: 1, Params: nil},
			{Tool: "clair", Priority: 2, Params: nil},
			{Tool: "docker-bench-security", Priority: 3, Params: nil},
		},
		"iac_security_assessment": {
			{Tool: "checkov", Priority: 1, Params: nil},
			{Tool: "terrascan", Priority: 2, Params: nil},
			{Tool: "trivy", Priority: 3, Params: nil},
		},
		"multi_cloud_assessment": {
			{Tool: "scout-suite", Priority: 1, Params: nil},
			{Tool: "prowler", Priority: 2, Params: nil},
			{Tool: "checkov", Priority: 3, Params: nil},
			{Tool: "terrascan", Priority: 4, Params: nil},
		},
		"bug_bounty_reconnaissance": {
			{Tool: "amass", Priority: 1, Params: map[string]string{"mode": "enum", "additional_args": "-passive"}},
			{Tool: "subfinder", Priority: 2, Params: map[string]string{"additional_args": "-silent"}},
			{Tool: "httpx", Priority: 3, Params: map[string]string{"additional_args": "-silent -tech-detect -status-code"}},
			{Tool: "katana", Priority: 4, Params: map[string]string{"additional_args": "-silent"}},
			{Tool: "gau", Priority: 5, Params: map[string]string{"additional_args": "--threads 5"}},
			{Tool: "arjun", Priority: 6, Params: map[string]string{"method": "GET,POST"}},
		},
		"bug_bounty_vulnerability_hunting": {
			{Tool: "nuclei", Priority: 1, Params: map[string]string{"severity": "critical,high"}},
			{Tool: "dalfox", Priority: 2, Params: nil},
			{Tool: "sqlmap", Priority: 3, Params: nil},
			{Tool: "ffuf", Priority: 4, Params: nil},
		},
		"bug_bounty_high_impact": {
			{Tool: "nuclei", Priority: 1, Params: map[string]string{"severity": "critical"}},
			{Tool: "sqlmap", Priority: 2, Params: map[string]string{"additional_args": "--batch"}},
			{Tool: "dalfox", Priority: 3, Params: nil},
		},
		"osint_passive_recon": {
			{Tool: "amass", Priority: 1, Params: map[string]string{"mode": "enum"}},
			{Tool: "subfinder", Priority: 2, Params: nil},
			{Tool: "theharvester", Priority: 3, Params: nil},
			{Tool: "httpx", Priority: 4, Params: map[string]string{"additional_args": "-silent"}},
		},
		"active_directory_assessment": {
			{Tool: "enum4linux", Priority: 1, Params: nil},
			{Tool: "netexec", Priority: 2, Params: nil},
			{Tool: "bloodhound", Priority: 3, Params: nil},
			{Tool: "responder", Priority: 4, Params: nil},
		},
		"wireless_assessment": {
			{Tool: "aircrack", Priority: 1, Params: nil},
			{Tool: "bettercap", Priority: 2, Params: nil},
		},
		"mobile_app_assessment": {
			{Tool: "trivy", Priority: 1, Params: nil},
			{Tool: "nuclei", Priority: 2, Params: nil},
		},
		"supply_chain_assessment": {
			{Tool: "trivy", Priority: 1, Params: nil},
			{Tool: "gitleaks", Priority: 2, Params: nil},
			{Tool: "trufflehog", Priority: 3, Params: nil},
		},
		"phishing_assessment": {
			{Tool: "theharvester", Priority: 1, Params: nil},
			{Tool: "httpx", Priority: 2, Params: nil},
		},
	}
}

var stealthToolIDs = map[string]struct{}{
	"amass": {}, "subfinder": {}, "httpx": {}, "nuclei": {},
}

// SelectPatternKey picks an attack pattern name from target type and objective.
func SelectPatternKey(targetType, objective string) string {
	obj := strings.ToLower(strings.TrimSpace(objective))
	switch obj {
	case "recon", "reconnaissance":
		if targetType == "api" {
			return "api_testing"
		}
		return "web_reconnaissance"
	case "vuln", "vulnerability", "vuln-hunt", "vuln_hunt":
		return "vulnerability_assessment"
	case "bugbounty-recon", "osint":
		return "bug_bounty_reconnaissance"
	case "bugbounty-high", "high-impact":
		return "bug_bounty_high_impact"
	case "business-logic", "file-upload":
		return "bug_bounty_vulnerability_hunting"
	case "ctf", "ctf-web":
		return "ctf_web_challenge"
	case "pwn", "ctf-pwn":
		return "ctf_pwn_challenge"
	case "forensics":
		return "ctf_forensics_challenge"
	case "crypto":
		return "ctf_crypto_challenge"
	case "binary", "exploit":
		return "binary_exploitation"
	case "ad", "active-directory":
		return "active_directory_assessment"
	case "supply-chain":
		return "supply_chain_assessment"
	case "wireless":
		return "wireless_assessment"
	case "mobile":
		return "mobile_app_assessment"
	case "phishing":
		return "phishing_assessment"
	}
	switch targetType {
	case "web":
		if obj == "comprehensive" {
			return "vulnerability_assessment"
		}
		return "web_reconnaissance"
	case "api":
		return "api_testing"
	case "ip":
		if obj == "comprehensive" {
			return "comprehensive_network_pentest"
		}
		return "network_discovery"
	case "cloud":
		if obj == "multi-cloud" {
			return "multi_cloud_assessment"
		}
		return "aws_security_assessment"
	case "binary":
		return "binary_exploitation"
	default:
		return "web_reconnaissance"
	}
}

// FilterStealthTools keeps only low-noise tool ids (max 4).
func FilterStealthTools(toolIDs []string) []string {
	out := make([]string, 0, 4)
	for _, id := range toolIDs {
		if _, ok := stealthToolIDs[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// FilterComprehensiveTools keeps tools with effectiveness above 0.7.
func FilterComprehensiveTools(eng *DecisionEngine, targetType string, toolIDs []string) []string {
	out := make([]string, 0, len(toolIDs))
	for _, id := range toolIDs {
		if eng.Score(targetType, id) > 0.7 {
			out = append(out, id)
		}
	}
	return out
}
