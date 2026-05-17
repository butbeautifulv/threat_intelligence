package decision

import "strings"

// OptimizeContext carries objective hints for parameter tuning.
type OptimizeContext struct {
	Stealth      bool
	Aggressive   bool
	Quick        bool
	Objective    string
}

// OptimizeParametersWithProfile applies HexStrike per-tool optimizers with profile awareness.
func (d *DecisionEngine) OptimizeParametersWithProfile(p TargetProfile, toolID string, params map[string]string, ctx OptimizeContext) map[string]string {
	out := make(map[string]string)
	for k, v := range params {
		out[k] = v
	}
	if out["target"] == "" {
		out["target"] = p.Target
	}
	switch toolID {
	case "nmap":
		optimizeNmapParams(p, ctx, out)
	case "nmap-advanced":
		optimizeNmapAdvancedParams(p, ctx, out)
	case "gobuster":
		optimizeGobusterParams(p, ctx, out)
	case "nuclei":
		optimizeNucleiParams(p, ctx, out)
	case "sqlmap":
		optimizeSQLMapParams(p, ctx, out)
	case "ffuf":
		optimizeFfufParams(p, ctx, out)
	case "hydra":
		optimizeHydraParams(p, out)
	case "rustscan":
		optimizeRustscanParams(ctx, out)
	case "masscan":
		optimizeMasscanParams(ctx, out)
	case "enum4linux-ng", "enum4linux":
		out["shares"] = "true"
		out["users"] = "true"
		out["groups"] = "true"
	case "autorecon":
		out["port_scans"] = "top-1000-ports"
		out["service_scans"] = "default"
	case "ghidra":
		if out["analysis_timeout"] == "" {
			out["analysis_timeout"] = "300"
		}
	case "pwntools":
		if out["exploit_type"] == "" {
			out["exploit_type"] = "local"
		}
	case "ropper":
		if out["gadget_type"] == "" {
			out["gadget_type"] = "rop"
		}
	case "angr":
		if out["analysis_type"] == "" {
			out["analysis_type"] = "symbolic"
		}
	case "prowler":
		if out["provider"] == "" {
			out["provider"] = "aws"
		}
		if out["output_format"] == "" {
			out["output_format"] = "json"
		}
	case "scout-suite":
		if out["provider"] == "" {
			out["provider"] = "aws"
		}
	case "kube-hunter":
		if out["report"] == "" {
			out["report"] = "json"
		}
	case "trivy":
		if out["scan_type"] == "" {
			out["scan_type"] = "image"
		}
		if out["severity"] == "" {
			out["severity"] = "HIGH,CRITICAL"
		}
	case "checkov":
		if out["output_format"] == "" {
			out["output_format"] = "json"
		}
	default:
		applyLegacyParameterDefaults(p.TargetType, toolID, out)
	}
	return out
}

func applyLegacyParameterDefaults(targetType, toolID string, out map[string]string) {
	switch toolID {
	case "httpx":
		if out["additional_args"] == "" {
			out["additional_args"] = "-silent"
		}
	case "feroxbuster":
		if out["additional_args"] == "" {
			out["additional_args"] = "-q"
		}
	case "nikto":
		if out["additional_args"] == "" {
			out["additional_args"] = "-Tuning 123bde"
		}
	case "wpscan":
		if out["additional_args"] == "" {
			out["additional_args"] = "--enumerate vp,vt,tt,u,m"
		}
	default:
		if toolID == "nmap" && out["scan_type"] == "" {
			out["scan_type"] = "-sV"
			out["additional_args"] = "-T4 -Pn"
		}
		if toolID == "nuclei" && out["templates"] == "" && targetType == "web" {
			out["templates"] = "cves/,misconfiguration/"
		}
	}
}

func optimizeNmapParams(p TargetProfile, ctx OptimizeContext, out map[string]string) {
	switch p.TargetType {
	case "web":
		if out["scan_type"] == "" {
			out["scan_type"] = "-sV -sC"
		}
		if out["ports"] == "" {
			out["ports"] = "80,443,8080,8443,8000,9000"
		}
	case "ip":
		if out["scan_type"] == "" {
			out["scan_type"] = "-sS -O"
		}
		if out["additional_args"] == "" {
			out["additional_args"] = "--top-ports 1000"
		}
	default:
		if out["scan_type"] == "" {
			out["scan_type"] = "-sV"
		}
	}
	args := out["additional_args"]
	if ctx.Stealth {
		out["additional_args"] = strings.TrimSpace(args + " -T2")
	} else if !strings.Contains(args, "-T") {
		out["additional_args"] = strings.TrimSpace(args + " -T4 -Pn")
	}
}

func optimizeNmapAdvancedParams(p TargetProfile, ctx OptimizeContext, out map[string]string) {
	if ctx.Stealth {
		out["scan_type"] = "-sS"
		out["timing"] = "T2"
	} else {
		out["scan_type"] = "-sS"
		out["timing"] = "T4"
		out["os_detection"] = "true"
		out["version_detection"] = "true"
	}
	switch p.TargetType {
	case "web":
		out["nse_scripts"] = "http-*,ssl-*"
	case "ip":
		out["nse_scripts"] = "default,discovery,safe"
	}
}

func optimizeGobusterParams(p TargetProfile, _ OptimizeContext, out map[string]string) {
	if out["mode"] == "" {
		out["mode"] = "dir"
	}
	ext := "-x html,php,txt,js"
	switch {
	case profileHasTech(p, "php"):
		ext = "-x php,html,txt,xml"
	case profileHasTech(p, "dotnet", "asp.net"):
		ext = "-x asp,aspx,html,txt"
	case profileHasTech(p, "java", "tomcat"):
		ext = "-x jsp,html,txt,xml"
	}
	if out["additional_args"] == "" {
		out["additional_args"] = ext + " -t 20"
	}
	if out["wordlist"] == "" {
		out["wordlist"] = "/usr/share/wordlists/dirb/common.txt"
	}
}

func optimizeNucleiParams(p TargetProfile, ctx OptimizeContext, out map[string]string) {
	if ctx.Quick {
		out["severity"] = "critical,high"
	} else if out["severity"] == "" {
		out["severity"] = "critical,high,medium"
	}
	var tags []string
	if profileHasTech(p, "wordpress") {
		tags = append(tags, "wordpress")
	}
	if profileHasTech(p, "drupal") {
		tags = append(tags, "drupal")
	}
	if profileHasTech(p, "joomla") {
		tags = append(tags, "joomla")
	}
	if len(tags) > 0 && out["tags"] == "" {
		out["tags"] = strings.Join(tags, ",")
	}
	if out["templates"] == "" && p.TargetType == "web" {
		out["templates"] = "cves/,misconfiguration/"
	}
}

func optimizeSQLMapParams(p TargetProfile, ctx OptimizeContext, out map[string]string) {
	switch {
	case profileHasTech(p, "php"):
		out["additional_args"] = "--dbms=mysql --batch"
	case profileHasTech(p, "dotnet", "asp.net"):
		out["additional_args"] = "--dbms=mssql --batch"
	default:
		if out["additional_args"] == "" {
			out["additional_args"] = "--batch --random-agent"
		}
	}
	if ctx.Aggressive {
		out["additional_args"] = strings.TrimSpace(out["additional_args"] + " --level=3 --risk=2")
	}
}

func optimizeFfufParams(p TargetProfile, ctx OptimizeContext, out map[string]string) {
	if p.TargetType == "api" {
		out["match_codes"] = "200,201,202,204,301,302,401,403"
	} else if out["match_codes"] == "" {
		out["match_codes"] = "200,204,301,302,307,401,403"
	}
	if ctx.Stealth {
		out["additional_args"] = "-t 10 -p 1"
	} else if out["additional_args"] == "" {
		out["additional_args"] = "-t 40"
	}
}

func optimizeHydraParams(_ TargetProfile, out map[string]string) {
	if out["service"] == "" {
		out["service"] = "ssh"
	}
	if out["additional_args"] == "" {
		out["additional_args"] = "-t 4 -w 30"
	}
}

func optimizeRustscanParams(ctx OptimizeContext, out map[string]string) {
	if ctx.Stealth {
		out["ulimit"] = "1000"
		out["batch_size"] = "500"
	} else if ctx.Aggressive {
		out["ulimit"] = "10000"
		out["batch_size"] = "8000"
	} else {
		out["ulimit"] = "5000"
		out["batch_size"] = "4500"
	}
	if ctx.Objective == "comprehensive" {
		out["scripts"] = "true"
	}
	if out["additional_args"] == "" {
		out["additional_args"] = "-- -sV"
	}
}

func optimizeMasscanParams(ctx OptimizeContext, out map[string]string) {
	if ctx.Stealth {
		out["rate"] = "100"
	} else if ctx.Aggressive {
		out["rate"] = "10000"
	} else if out["rate"] == "" {
		out["rate"] = "1000"
	}
	if out["ports"] == "" {
		out["ports"] = "1-65535"
	}
	out["banners"] = "true"
}
