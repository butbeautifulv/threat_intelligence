package recovery

import (
	"regexp"
	"strings"
	"time"
)

// ErrorType classifies tool failures (HexStrike ErrorType enum values).
type ErrorType string

const (
	TypeTimeout              ErrorType = "timeout"
	TypeToolNotFound         ErrorType = "tool_not_found"
	TypeRateLimited          ErrorType = "rate_limited"
	TypePermissionDenied     ErrorType = "permission_denied"
	TypeNetworkUnreachable   ErrorType = "network_unreachable"
	TypeInvalidParams        ErrorType = "invalid_parameters"
	TypeResourceExhausted    ErrorType = "resource_exhausted"
	TypeAuthenticationFailed ErrorType = "authentication_failed"
	TypeTargetUnreachable    ErrorType = "target_unreachable"
	TypeParsing              ErrorType = "parsing_error"
	TypeUnknown              ErrorType = "unknown"
)

// Handler suggests recovery actions for failed tool runs.
type Handler struct {
	patterns     []*regexp.Regexp
	types        []ErrorType
	alternatives map[string][]string
}

func Default() *Handler {
	patterns := []string{
		`(?i)timeout|timed out|connection timeout|read timeout|operation timed out|command timeout`,
		`(?i)permission denied|access denied|forbidden|not authorized|sudo required|root required|insufficient privileges`,
		`(?i)network unreachable|host unreachable|no route to host|connection refused|connection reset|network error`,
		`(?i)rate limit|too many requests|throttled|429|request limit exceeded|quota exceeded`,
		`(?i)command not found|no such file or directory|no such file|executable not found|binary not found`,
		`(?i)invalid argument|invalid option|unknown option|bad parameter|invalid parameter|syntax error`,
		`(?i)out of memory|memory error|disk full|no space left|resource temporarily unavailable|too many open files`,
		`(?i)authentication failed|login failed|invalid credentials|invalid token|expired token`,
		`(?i)target unreachable|target not responding|target down|host not found|dns resolution failed`,
		`(?i)parse error|parsing failed|invalid format|malformed|json decode error|xml parse error|invalid json`,
	}
	types := []ErrorType{
		TypeTimeout, TypePermissionDenied, TypeNetworkUnreachable, TypeRateLimited, TypeToolNotFound,
		TypeInvalidParams, TypeResourceExhausted, TypeAuthenticationFailed, TypeTargetUnreachable, TypeParsing,
	}
	var compiled []*regexp.Regexp
	for _, p := range patterns {
		compiled = append(compiled, regexp.MustCompile(p))
	}
	return &Handler{
		patterns: compiled,
		types:    types,
		alternatives: map[string][]string{
			"nmap":        {"rustscan", "masscan"},
			"rustscan":    {"nmap", "masscan"},
			"masscan":     {"nmap", "rustscan"},
			"gobuster":    {"feroxbuster", "dirsearch", "ffuf"},
			"feroxbuster": {"gobuster", "dirsearch", "ffuf"},
			"dirsearch":   {"gobuster", "feroxbuster"},
			"ffuf":        {"gobuster", "feroxbuster"},
			"nuclei":      {"jaeles", "nikto"},
			"jaeles":      {"nuclei", "nikto"},
			"nikto":       {"nuclei", "jaeles"},
			"katana":      {"gau", "waybackurls"},
			"gau":         {"katana", "waybackurls"},
			"arjun":       {"paramspider", "x8"},
			"paramspider": {"arjun", "x8"},
			"sqlmap":      {"sqlmap"},
			"dalfox":      {"dalfox"},
			"subfinder":   {"amass", "subfinder"},
			"amass":       {"subfinder"},
		},
	}
}

// Alternatives returns tool → fallback binary names (read-only diagnostics).
func (h *Handler) Alternatives() map[string][]string {
	out := make(map[string][]string, len(h.alternatives))
	for k, v := range h.alternatives {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

func (h *Handler) Classify(msg string) ErrorType {
	for i, re := range h.patterns {
		if re.MatchString(msg) {
			return h.types[i]
		}
	}
	return TypeUnknown
}

func (h *Handler) Recoverable(t ErrorType) bool {
	switch t {
	case TypeTimeout, TypeToolNotFound, TypeRateLimited, TypeNetworkUnreachable, TypeTargetUnreachable, TypeInvalidParams:
		return true
	default:
		return false
	}
}

func (h *Handler) SuggestAlternative(tool string, t ErrorType) string {
	bin := toolBinary(tool)
	alts, ok := h.alternatives[bin]
	if !ok {
		return ""
	}
	switch t {
	case TypeTimeout, TypeToolNotFound, TypeNetworkUnreachable:
		if len(alts) > 0 {
			return alts[0]
		}
	}
	return ""
}

func (h *Handler) AdjustParams(tool string, t ErrorType, params map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range params {
		out[k] = v
	}
	bin := toolBinary(tool)
	switch t {
	case TypeTimeout:
		switch bin {
		case "nmap":
			out["additional_args"] = strings.TrimSpace(out["additional_args"] + " -T2")
		case "gobuster", "feroxbuster", "ffuf":
			if out["threads"] == "" {
				out["threads"] = "10"
			}
			if out["timeout"] == "" {
				out["timeout"] = "30"
			}
		case "nuclei":
			if out["additional_args"] == "" {
				out["additional_args"] = "-c 10"
			}
		default:
			if out["timeout"] == "" {
				out["timeout"] = "60"
			}
			if out["threads"] == "" {
				out["threads"] = "5"
			}
		}
		out["_reduce_scope"] = "true"
	case TypeRateLimited:
		switch bin {
		case "nmap":
			out["additional_args"] = strings.TrimSpace(out["additional_args"] + " -T1")
		case "gobuster", "feroxbuster", "ffuf":
			out["threads"] = "5"
			out["delay"] = "1"
		case "nuclei":
			if out["additional_args"] == "" {
				out["additional_args"] = "-c 5 -rl 10"
			}
		default:
			out["delay"] = "2"
			out["threads"] = "3"
		}
	case TypeInvalidParams:
		if out["additional_args"] != "" {
			out["additional_args"] = strings.TrimSpace(out["additional_args"] + " --help")
		}
	case TypeResourceExhausted:
		switch bin {
		case "nmap":
			out["additional_args"] = strings.TrimSpace(out["additional_args"] + " --max-parallelism 10")
		case "gobuster", "feroxbuster", "ffuf", "nuclei":
			out["threads"] = "5"
		default:
			if out["threads"] != "" {
				out["threads"] = "3"
			}
		}
	}
	return out
}

// BackoffDelay returns sleep duration for attempt (1-based).
func (h *Handler) BackoffDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	sec := 5
	for i := 1; i < attempt; i++ {
		sec *= 2
		if sec > 60 {
			sec = 60
			break
		}
	}
	return time.Duration(sec) * time.Second
}

// MaxRetries returns bounded retry count for recoverable errors.
func (h *Handler) MaxRetries(t ErrorType) int {
	switch t {
	case TypeTimeout, TypeRateLimited, TypeNetworkUnreachable:
		return 3
	case TypeToolNotFound:
		return 1
	default:
		return 2
	}
}

func toolBinary(tool string) string {
	if idx := strings.Index(tool, "_"); idx > 0 {
		return tool[:idx]
	}
	return tool
}
