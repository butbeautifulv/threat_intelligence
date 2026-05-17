package findings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	domainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
)

var (
	masscanOpenLine = regexp.MustCompile(`(?i)^open\s+(tcp|udp)\s+(\d+)\s+(\S+)`)
	grepablePorts   = regexp.MustCompile(`(?i)Ports:\s*(.+)$`)
)

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
