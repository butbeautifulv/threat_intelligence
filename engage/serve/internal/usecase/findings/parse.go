package findings

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

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

// ParseToolOutput extracts structured findings from tool stdout.
func ParseToolOutput(toolName, target, output string) []domainreport.Finding {
	low := strings.ToLower(toolName)
	var parsed []domainreport.Finding
	switch {
	case strings.Contains(low, "nuclei"):
		parsed = parseNuclei(target, toolName, output)
	case strings.Contains(low, "nmap"):
		parsed = parseNmap(target, toolName, output)
	case strings.Contains(low, "masscan"):
		parsed = parseMasscan(target, toolName, output)
	case strings.Contains(low, "ffuf"):
		parsed = parseFfuf(target, toolName, output)
	case strings.Contains(low, "sqlmap"):
		parsed = parseSqlmap(target, toolName, output)
	case strings.Contains(low, "wpscan"):
		parsed = parseWpscan(target, toolName, output)
	default:
		parsed = parseGeneric(target, toolName, output)
	}
	return DedupeFindings(parsed)
}

var (
	masscanOpenLine  = regexp.MustCompile(`(?i)^open\s+(tcp|udp)\s+(\d+)\s+(\S+)`)
	grepablePorts    = regexp.MustCompile(`(?i)Ports:\s*(.+)$`)
	wpscanTitleLine  = regexp.MustCompile(`^\s*\[!]\s+(.*)$`)
	signatureSpaceRe = regexp.MustCompile(`\s+`)
)

func normalizeSignature(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return signatureSpaceRe.ReplaceAllString(s, " ")
}

func parseMasscan(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "Ports:") {
			fields := grepablePorts.FindStringSubmatch(line)
			host := extractGrepableHost(line)
			if len(fields) != 2 {
				continue
			}
			for _, frag := range strings.Split(fields[1], ",") {
				frag = strings.TrimSpace(frag)
				openPortFinding(&out, host, target, tool, frag, line)
			}
			continue
		}
		m := masscanOpenLine.FindStringSubmatch(line)
		if len(m) == 4 {
			portNum, _ := strconv.Atoi(m[2])
			ip := strings.TrimSpace(m[3])
			useTarget := ip
			if useTarget == "" {
				useTarget = target
			}
			out = append(out, domainreport.Finding{
				Title:       fmt.Sprintf("open %s/%s", m[2], strings.ToLower(m[1])),
				Severity:    domainreport.SeverityInfo,
				Description: fmt.Sprintf("masscan detected open port %d (%s)", portNum, strings.ToUpper(m[1])),
				Target:      useTarget,
				Tool:        tool,
				Evidence:    line,
			})
		}
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func extractGrepableHost(line string) string {
	idx := strings.Index(line, "Host:")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(line[idx+len("Host:"):])
	if p := strings.Index(rest, "()"); p >= 0 {
		host := strings.TrimSpace(rest[:p])
		host = strings.Trim(host, "()")
		return strings.TrimSpace(host)
	}
	fields := strings.Fields(rest)
	if len(fields) > 0 {
		return strings.Trim(fields[0], "()")
	}
	return strings.TrimSpace(rest)
}

func openPortFinding(dst *[]domainreport.Finding, host, targetFallback, tool, portFrag, evidence string) {
	// e.g. 22/open/tcp//ssh/, 443/open/tcp//https/
	frag := strings.TrimSpace(strings.ToLower(portFrag))
	fields := strings.Split(frag, "/")
	if len(fields) < 2 {
		return
	}
	if fields[1] != "open" {
		return
	}
	proto := "tcp"
	if len(fields) > 2 && fields[2] != "" {
		proto = fields[2]
	}
	useTarget := host
	if useTarget == "" {
		useTarget = targetFallback
	}
	title := fmt.Sprintf("open %s/%s", fields[0], proto)
	*dst = append(*dst, domainreport.Finding{
		Title:       title,
		Severity:    domainreport.SeverityInfo,
		Description: frag,
		Target:      useTarget,
		Tool:        tool,
		Evidence:    evidence,
	})
}

func parseWpscan(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	trim := strings.TrimSpace(output)
	if trim != "" && trim[0] == '{' {
		type vulnStub struct {
			Title       string `json:"title"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		var doc struct {
			Vulnerabilities []vulnStub `json:"vulnerabilities"`
			Interesting     []struct {
				URL   string `json:"url"`
				Found string `json:"found_by"`
			} `json:"interesting_findings"`
		}
		if err := json.Unmarshal([]byte(trim), &doc); err == nil {
			for _, v := range doc.Vulnerabilities {
				title := strings.TrimSpace(v.Title)
				if title == "" {
					title = strings.TrimSpace(v.Name)
				}
				if title == "" {
					continue
				}
				desc := strings.TrimSpace(v.Description)
				out = append(out, domainreport.Finding{
					Title:       "wpscan: " + title,
					Severity:    domainreport.SeverityMedium,
					Description: desc,
					Target:      target,
					Tool:        tool,
				})
			}
			for _, it := range doc.Interesting {
				if it.URL == "" {
					continue
				}
				out = append(out, domainreport.Finding{
					Title:       "wpscan: " + strings.TrimSpace(it.URL),
					Severity:    domainreport.SeverityInfo,
					Description: strings.TrimSpace(it.Found),
					Target:      target,
					Tool:        tool,
					Evidence:    strings.TrimSpace(it.URL),
				})
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if m := wpscanTitleLine.FindStringSubmatch(line); len(m) == 2 {
			msg := strings.TrimSpace(m[1])
			if msg == "" {
				continue
			}
			low := strings.ToLower(msg)
			sev := domainreport.SeverityLow
			if strings.Contains(low, "critical") {
				sev = domainreport.SeverityCritical
			} else if strings.Contains(low, "high") {
				sev = domainreport.SeverityHigh
			} else if strings.Contains(low, "medium") {
				sev = domainreport.SeverityMedium
			}
			out = append(out, domainreport.Finding{
				Title:       "wpscan: " + msg,
				Severity:    sev,
				Description: msg,
				Target:      target,
				Tool:        tool,
				Evidence:    line,
			})
		}
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func parseNuclei(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var rec struct {
			Info struct {
				Name     string `json:"name"`
				Severity string `json:"severity"`
			} `json:"info"`
			MatcherName string `json:"matcher-name"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		title := rec.Info.Name
		if title == "" {
			title = rec.MatcherName
		}
		if title == "" {
			continue
		}
		out = append(out, domainreport.Finding{
			Title:       title,
			Severity:    mapSeverity(rec.Info.Severity),
			Description: title,
			Target:      target,
			Tool:        tool,
			Evidence:    line,
		})
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func parseNmap(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "/tcp") && !strings.Contains(line, "/udp") {
			continue
		}
		if !strings.Contains(strings.ToLower(line), "open") {
			continue
		}
		out = append(out, domainreport.Finding{
			Title:       "open port",
			Severity:    domainreport.SeverityInfo,
			Description: strings.TrimSpace(line),
			Target:      target,
			Tool:        tool,
			Evidence:    line,
		})
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func parseFfuf(target, tool, output string) []domainreport.Finding {
	trim := strings.TrimSpace(output)
	if trim == "" {
		return nil
	}
	if trim[0] == '{' {
		var doc struct {
			Results []struct {
				URL    string            `json:"url"`
				Status int               `json:"status"`
				Input  map[string]string `json:"input"`
			} `json:"results"`
		}
		if err := json.Unmarshal([]byte(trim), &doc); err == nil && len(doc.Results) > 0 {
			var out []domainreport.Finding
			for _, r := range doc.Results {
				title := r.URL
				if title == "" {
					for _, v := range r.Input {
						title = v
						break
					}
				}
				if title == "" {
					continue
				}
				out = append(out, domainreport.Finding{
					Title:       "ffuf: " + title,
					Severity:    ffufSeverity(r.Status),
					Description: fmt.Sprintf("status %d", r.Status),
					Target:      target,
					Tool:        tool,
					Evidence:    fmt.Sprintf(`{"url":%q,"status":%d}`, title, r.Status),
				})
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	var out []domainreport.Finding
	statusRe := regexp.MustCompile(`(?i)status:\s*(\d+)`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := statusRe.FindStringSubmatch(line)
		if len(m) == 2 {
			code, _ := strconv.Atoi(m[1])
			out = append(out, domainreport.Finding{
				Title:       "ffuf match",
				Severity:    ffufSeverity(code),
				Description: line,
				Target:      target,
				Tool:        tool,
				Evidence:    line,
			})
		}
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func ffufSeverity(status int) domainreport.Severity {
	switch {
	case status >= 500:
		return domainreport.SeverityHigh
	case status == 401 || status == 403:
		return domainreport.SeverityMedium
	case status >= 200 && status < 300:
		return domainreport.SeverityLow
	default:
		return domainreport.SeverityInfo
	}
}

func parseSqlmap(target, tool, output string) []domainreport.Finding {
	var out []domainreport.Finding
	low := strings.ToLower(output)
	if strings.Contains(low, "sqlmap identified the following injection point") {
		out = append(out, domainreport.Finding{
			Title:       "SQL injection point identified",
			Severity:    domainreport.SeverityHigh,
			Description: "sqlmap identified injectable parameter(s)",
			Target:      target,
			Tool:        tool,
		})
	}
	paramRe := regexp.MustCompile(`(?m)^Parameter:\s*(.+)$`)
	typeRe := regexp.MustCompile(`(?m)^\s+Type:\s*(.+)$`)
	titleRe := regexp.MustCompile(`(?m)^\s+Title:\s*(.+)$`)
	params := paramRe.FindAllStringSubmatch(output, -1)
	types := typeRe.FindAllStringSubmatch(output, -1)
	titles := titleRe.FindAllStringSubmatch(output, -1)
	for i, pm := range params {
		desc := strings.TrimSpace(pm[1])
		if i < len(types) {
			desc += " (" + strings.TrimSpace(types[i][1]) + ")"
		}
		title := "sqlmap: " + strings.TrimSpace(pm[1])
		if i < len(titles) && titles[i][1] != "" {
			title = "sqlmap: " + strings.TrimSpace(titles[i][1])
		}
		out = append(out, domainreport.Finding{
			Title:       title,
			Severity:    domainreport.SeverityHigh,
			Description: desc,
			Target:      target,
			Tool:        tool,
		})
	}
	if strings.Contains(low, "is vulnerable") {
		out = append(out, domainreport.Finding{
			Title:       "sqlmap: target is vulnerable",
			Severity:    domainreport.SeverityHigh,
			Description: "sqlmap reported vulnerability",
			Target:      target,
			Tool:        tool,
		})
	}
	if len(out) == 0 {
		return parseGeneric(target, tool, output)
	}
	return out
}

func parseGeneric(target, tool, output string) []domainreport.Finding {
	indicators := []struct {
		word     string
		severity domainreport.Severity
	}{
		{"CRITICAL", domainreport.SeverityCritical},
		{"HIGH", domainreport.SeverityHigh},
		{"MEDIUM", domainreport.SeverityMedium},
		{"VULNERABILITY", domainreport.SeverityHigh},
		{"SQL injection", domainreport.SeverityHigh},
		{"XSS", domainreport.SeverityMedium},
	}
	var out []domainreport.Finding
	low := strings.ToLower(output)
	for _, ind := range indicators {
		if strings.Contains(low, strings.ToLower(ind.word)) {
			out = append(out, domainreport.Finding{
				Title:       ind.word + " indicator",
				Severity:    ind.severity,
				Description: "matched output pattern",
				Target:      target,
				Tool:        tool,
			})
		}
	}
	return out
}

func mapSeverity(s string) domainreport.Severity {
	switch strings.ToLower(s) {
	case "critical":
		return domainreport.SeverityCritical
	case "high":
		return domainreport.SeverityHigh
	case "medium":
		return domainreport.SeverityMedium
	case "low":
		return domainreport.SeverityLow
	default:
		return domainreport.SeverityInfo
	}
}

// Count returns total findings across tool results.
func Count(all []domainreport.Finding) int {
	return len(all)
}
