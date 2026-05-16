package recovery

import (
	"regexp"
	"strings"
)

// ErrorType classifies tool failures.
type ErrorType string

const (
	TypeTimeout    ErrorType = "timeout"
	TypeNotFound   ErrorType = "not_found"
	TypeRateLimit  ErrorType = "rate_limit"
	TypePermission ErrorType = "permission"
	TypeUnknown    ErrorType = "unknown"
)

// Handler suggests recovery actions for failed tool runs.
type Handler struct {
	patterns     []*regexp.Regexp
	types        []ErrorType
	alternatives map[string]map[ErrorType]string
}

func Default() *Handler {
	patterns := []string{
		`(?i)timeout|timed out`,
		`(?i)command not found|no such file|not found`,
		`(?i)rate limit|too many requests|429`,
		`(?i)permission denied|access denied|forbidden`,
	}
	types := []ErrorType{TypeTimeout, TypeNotFound, TypeRateLimit, TypePermission}
	var compiled []*regexp.Regexp
	for _, p := range patterns {
		compiled = append(compiled, regexp.MustCompile(p))
	}
	return &Handler{
		patterns: compiled,
		types:    types,
		alternatives: map[string]map[ErrorType]string{
			"nuclei":     {TypeNotFound: "nikto"},
			"gobuster":   {TypeTimeout: "feroxbuster"},
			"feroxbuster": {TypeTimeout: "gobuster"},
			"nmap":       {TypeTimeout: "rustscan"},
			"rustscan":   {TypeNotFound: "nmap"},
		},
	}
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
	case TypeTimeout, TypeNotFound, TypeRateLimit:
		return true
	default:
		return false
	}
}

func (h *Handler) SuggestAlternative(tool string, t ErrorType) string {
	bin := toolBinary(tool)
	if m, ok := h.alternatives[bin]; ok {
		if alt, ok := m[t]; ok {
			return alt
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
		case "feroxbuster", "gobuster":
			if out["threads"] == "" {
				out["threads"] = "10"
			}
		}
	case TypeRateLimit:
		out["additional_args"] = strings.TrimSpace(out["additional_args"] + " --delay 1")
	}
	return out
}

func toolBinary(tool string) string {
	if idx := strings.Index(tool, "_"); idx > 0 {
		return tool[:idx]
	}
	return tool
}
