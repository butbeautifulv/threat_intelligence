package security

import (
	"net"
	"net/url"
	"strings"
)

// TargetGuardMode controls SSRF-style blocking for tool targets.
type TargetGuardMode string

const (
	TargetGuardOff   TargetGuardMode = "off"
	TargetGuardWarn  TargetGuardMode = "warn"
	TargetGuardBlock TargetGuardMode = "block"
)

// ParseTargetGuardMode reads ENGAGE_TARGET_GUARD (off|warn|block).
// Default block when ENGAGE_ENV=prod unless explicitly off.
func ParseTargetGuardMode(getenv func(string) string) TargetGuardMode {
	raw := strings.ToLower(strings.TrimSpace(getenv("ENGAGE_TARGET_GUARD")))
	switch raw {
	case "off", "0", "false":
		return TargetGuardOff
	case "warn", "log":
		return TargetGuardWarn
	case "block", "1", "true":
		return TargetGuardBlock
	}
	if strings.EqualFold(strings.TrimSpace(getenv("ENGAGE_ENV")), "prod") {
		return TargetGuardBlock
	}
	return TargetGuardOff
}

// CheckTarget returns a reason if the target must not be scanned (cloud metadata, loopback, RFC1918).
// Intended for agentic tool runs in secured infra — operators scanning internal lab targets set ENGAGE_TARGET_GUARD=off.
func CheckTarget(target string) (blocked bool, reason string) {
	host := extractHost(target)
	if host == "" {
		return false, ""
	}
	if strings.EqualFold(host, "localhost") {
		return true, "localhost targets are blocked"
	}
	ip := net.ParseIP(host)
	if ip == nil {
		// hostname — allow; DNS resolution happens in tools
		return false, ""
	}
	if ip.IsLoopback() {
		return true, "loopback targets are blocked"
	}
	if ip.IsLinkLocalUnicast() || ip.Equal(net.IPv4(169, 254, 169, 254)) {
		return true, "link-local / metadata targets are blocked"
	}
	if ip.IsPrivate() {
		return true, "RFC1918 targets are blocked in guarded mode"
	}
	return false, ""
}

func extractHost(target string) string {
	t := strings.TrimSpace(target)
	if t == "" {
		return ""
	}
	if strings.Contains(t, "://") {
		u, err := url.Parse(t)
		if err != nil || u.Host == "" {
			return ""
		}
		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return strings.ToLower(u.Host)
		}
		return strings.ToLower(host)
	}
	if h, _, err := net.SplitHostPort(t); err == nil {
		return strings.ToLower(h)
	}
	// bare IP or hostname
	if i := strings.Index(t, "/"); i >= 0 {
		t = t[:i]
	}
	return strings.ToLower(strings.TrimSpace(t))
}
