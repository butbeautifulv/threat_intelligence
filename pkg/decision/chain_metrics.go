package decision

// executionTimeEstimates mirrors HexStrike create_attack_chain time_estimates (seconds).
var executionTimeEstimates = map[string]int{
	"nmap": 120, "gobuster": 300, "nuclei": 180, "nikto": 240,
	"sqlmap": 600, "ffuf": 200, "hydra": 900, "amass": 300,
	"ghidra": 300, "radare2": 180, "gdb": 120, "gdb-peda": 150,
	"angr": 600, "pwntools": 240, "ropper": 120, "one-gadget": 60,
	"checksec": 30, "pwninit": 60, "libc-database": 90,
	"prowler": 600, "scout-suite": 480, "cloudmapper": 300, "pacu": 420,
	"trivy": 180, "clair": 240, "kube-hunter": 300, "kube-bench": 120,
	"docker-bench-security": 180, "falco": 120, "checkov": 240, "terrascan": 200,
	"httpx": 60, "katana": 180, "rustscan": 90, "masscan": 120,
	"enum4linux": 240, "enum4linux-ng": 300, "autorecon": 900, "responder": 180,
	"dalfox": 240, "jaeles": 180, "wpscan": 300, "arjun": 120,
}

// ExecutionTimeEstimate returns estimated seconds for a tool step.
func ExecutionTimeEstimate(toolID string) int {
	if sec, ok := executionTimeEstimates[toolID]; ok {
		return sec
	}
	return 180
}

// ExpectedOutcome returns a default outcome label for a tool step.
func ExpectedOutcome(toolID string) string {
	return "Discover vulnerabilities using " + toolID
}

// StepSuccessProbability combines effectiveness and confidence into step probability.
func StepSuccessProbability(effectiveness, confidence float64) float64 {
	return effectiveness * confidence
}
