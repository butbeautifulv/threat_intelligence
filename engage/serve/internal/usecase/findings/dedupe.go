package findings

import (
	"regexp"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

var signatureSpaceRe = regexp.MustCompile(`\s+`)

// DedupeFindings merges findings that share (target, tool, normalized signature).
// Signature prefers Title; if empty, falls back to Description then Evidence preview.
func DedupeFindings(in []domainreport.Finding) []domainreport.Finding {
	if len(in) < 2 {
		return append([]domainreport.Finding(nil), in...)
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]domainreport.Finding, 0, len(in))
	for _, f := range in {
		key := findingDedupKey(f)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, f)
	}
	return out
}

func findingDedupKey(f domainreport.Finding) string {
	sig := strings.TrimSpace(f.Title)
	if sig == "" {
		sig = strings.TrimSpace(f.Description)
	}
	if sig == "" {
		sig = evidenceSignature(f.Evidence)
	}
	tgt := strings.ToLower(strings.TrimSpace(f.Target))
	tool := strings.ToLower(strings.TrimSpace(f.Tool))
	sig = normalizeSignature(sig)
	return tgt + "\x00" + tool + "\x00" + sig
}

func evidenceSignature(ev string) string {
	ev = strings.TrimSpace(ev)
	if ev == "" {
		return ""
	}
	runes := []rune(ev)
	if len(runes) > 200 {
		ev = string(runes[:200])
	}
	return ev
}

func normalizeSignature(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return signatureSpaceRe.ReplaceAllString(s, " ")
}
