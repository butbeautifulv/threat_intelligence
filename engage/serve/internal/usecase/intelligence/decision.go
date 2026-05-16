package intelligence

// DecisionEngine scores tools per target type (port of HexStrike IntelligentDecisionEngine tables).
type DecisionEngine struct {
	effectiveness map[string]map[string]float64
}

func DefaultDecisionEngine() *DecisionEngine {
	return &DecisionEngine{effectiveness: defaultEffectiveness()}
}

func defaultEffectiveness() map[string]map[string]float64 {
	return map[string]map[string]float64{
		"web": {
			"nmap": 0.8, "gobuster": 0.9, "nuclei": 0.95, "nikto": 0.85,
			"httpx": 0.85, "ffuf": 0.9, "feroxbuster": 0.85, "wpscan": 0.95,
		},
		"api": {
			"nuclei": 0.9, "ffuf": 0.85, "httpx": 0.9, "arjun": 0.95,
		},
		"ip": {
			"nmap": 0.95, "rustscan": 0.9, "masscan": 0.92, "nuclei": 0.85,
		},
		"cloud": {
			"prowler": 0.95, "trivy": 0.9, "scout-suite": 0.92, "kube-hunter": 0.9,
		},
		"unknown": {
			"nmap": 0.7, "httpx": 0.7, "subfinder": 0.75, "nuclei": 0.8,
		},
	}
}

// RankTools returns tool ids sorted by effectiveness for targetType.
func (d *DecisionEngine) RankTools(targetType string, candidates []string) []string {
	table, ok := d.effectiveness[targetType]
	if !ok {
		table = d.effectiveness["unknown"]
	}
	type scored struct {
		id    string
		score float64
	}
	var list []scored
	for _, id := range candidates {
		score := table[id]
		if score == 0 {
			score = 0.5
		}
		list = append(list, scored{id, score})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = s.id
	}
	return out
}

// OptimizeParameters applies tool-specific defaults from the decision engine.
func (d *DecisionEngine) OptimizeParameters(targetType, toolID string, params map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range params {
		out[k] = v
	}
	switch toolID {
	case "nmap":
		if out["scan_type"] == "" {
			out["scan_type"] = "-sV"
		}
		if out["additional_args"] == "" {
			out["additional_args"] = "-T4 -Pn"
		}
	case "nuclei":
		if out["templates"] == "" && targetType == "web" {
			out["templates"] = "cves/,misconfiguration/"
		}
	case "httpx":
		if out["additional_args"] == "" {
			out["additional_args"] = "-silent"
		}
	}
	return out
}
