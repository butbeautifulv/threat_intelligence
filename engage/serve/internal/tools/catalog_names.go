package tools

// BinaryToCatalog maps short tool ids (decision engine / legacy) to catalog entry names.
var BinaryToCatalog = map[string]string{
	"nmap":        "nmap_scan",
	"nuclei":      "nuclei_scan",
	"httpx":       "httpx_probe",
	"subfinder":   "subfinder_scan",
	"trivy":       "trivy_scan",
	"gobuster":    "gobuster_scan",
	"nikto":       "nikto_scan",
	"rustscan":    "rustscan_fast_scan",
	"feroxbuster": "feroxbuster_scan",
	"ffuf":        "ffuf_scan",
	"sqlmap":      "sqlmap_scan",
	"wpscan":      "wpscan_scan",
	"hydra":       "hydra_attack",
	"amass":       "amass_scan",
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
