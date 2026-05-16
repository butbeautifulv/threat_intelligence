package tools

// BinaryToCatalog maps short tool ids (decision engine / legacy) to catalog entry names.
var BinaryToCatalog = map[string]string{
	"nmap":                  "nmap_scan",
	"nmap-advanced":         "nmap_advanced_scan",
	"nuclei":                "nuclei_scan",
	"httpx":                 "httpx_probe",
	"subfinder":             "subfinder_scan",
	"trivy":                 "trivy_scan",
	"gobuster":              "gobuster_scan",
	"nikto":                 "nikto_scan",
	"rustscan":              "rustscan_fast_scan",
	"feroxbuster":           "feroxbuster_scan",
	"ffuf":                  "ffuf_scan",
	"sqlmap":                "sqlmap_scan",
	"wpscan":                "wpscan_analyze",
	"hydra":                 "hydra_attack",
	"amass":                 "amass_scan",
	"katana":                "katana_scan",
	"gau":                   "gau_discovery",
	"arjun":                 "arjun_scan",
	"paramspider":           "paramspider_scan",
	"dalfox":                "dalfox_xss_scan",
	"masscan":               "masscan_high_speed",
	"enum4linux":            "enum4linux_scan",
	"enum4linux-ng":         "enum4linux_ng_advanced",
	"smbmap":                "smbmap_scan",
	"autorecon":             "autorecon_comprehensive",
	"responder":             "responder_poison",
	"prowler":               "prowler_scan",
	"scout-suite":           "scout_suite_assessment",
	"cloudmapper":           "cloudmapper_analysis",
	"kube-bench":            "kube_bench_cis",
	"kube-hunter":           "kube_hunter_scan",
	"clair":                 "clair_vulnerability_scan",
	"docker-bench-security": "docker_bench_security_scan",
	"checkov":               "checkov_iac_scan",
	"terrascan":             "terrascan_scan",
	"checksec":              "checksec_analyze",
	"ghidra":                "ghidra_analysis",
	"radare2":               "radare2_scan",
	"gdb":                   "gdb_peda_debug",
	"strings":               "strings_extract",
	"exiftool":              "exiftool_extract",
	"binwalk":               "binwalk_analyze",
	"steghide":              "steghide_analysis",
	"theharvester":          "theharvester_scan",
	"netexec":               "netexec_scan",
	"bloodhound":            "bloodhound_ingest",
	"gitleaks":              "gitleaks_scan",
	"trufflehog":            "trufflehog_scan",
	"aircrack":              "aircrack_crack",
	"bettercap":             "bettercap_attack",
}

// ResolveCatalogName returns the catalog tool name for a short id or passes through if already a catalog name.
func ResolveCatalogName(id string, reg *Registry) string {
	if reg != nil {
		if _, ok := reg.Get(id); ok {
			return id
		}
	}
	if name, ok := BinaryToCatalog[id]; ok {
		if reg == nil {
			return name
		}
		if _, ok := reg.Get(name); ok {
			return name
		}
	}
	return id
}

// ResolveCatalogNames maps a list of tool ids to catalog names (skips unknown).
func ResolveCatalogNames(ids []string, reg *Registry) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{})
	for _, id := range ids {
		name := ResolveCatalogName(id, reg)
		if reg != nil {
			if _, ok := reg.Get(name); !ok {
				continue
			}
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
