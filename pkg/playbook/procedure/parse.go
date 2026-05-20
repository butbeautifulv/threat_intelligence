package procedure

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
)

var (
	frontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	sectionRe     = regexp.MustCompile(`(?m)^##\s+(.+)$`)
	stepRe        = regexp.MustCompile(`(?m)^###\s+Step\s+(\d+):\s*(.+)$`)
	toolTokenRe   = regexp.MustCompile(`(?i)\b(nmap|nuclei|masscan|nikto|sqlmap|gobuster|ffuf|feroxbuster|burp|wireshark|volatility|yara|sigma|splunk|elastic|misp|theharvester|amass|subfinder|httpx|wpscan|metasploit|bloodhound|hashcat|john|hydra|responder|impacket|shodan|censys|trivy|grype|kubesec|falco|osquery|velociraptor|autopsy|dcfldd|dd)\b`)
)

// ParseSkillMarkdown extracts a ProcedureSpec from SKILL.md content.
func ParseSkillMarkdown(id, subdomain string, attackIDs, nistCSF []string, raw string) domain.ProcedureSpec {
	body := raw
	if m := frontmatterRe.FindStringSubmatch(raw); len(m) > 1 {
		body = raw[len(m[0]):]
	}
	sections := splitSections(body)
	when := extractBullets(sections["when to use"])
	prereq := extractBullets(sections["prerequisites"])
	workflow := sections["workflow"]
	if workflow == "" {
		workflow = sections["detection workflow"]
	}
	steps := extractSteps(workflow)
	scenarios := extractBullets(sections["scenarios"])
	toolsBlock := sections["tools & systems"]
	if toolsBlock == "" {
		toolsBlock = sections["tools"]
	}
	mentions := collectMentions(steps, toolsBlock)
	return domain.ProcedureSpec{
		ID:            id,
		Subdomain:     subdomain,
		AttackIDs:     attackIDs,
		NISTCSF:       nistCSF,
		WhenToUse:     when,
		Prerequisites: prereq,
		Steps:         steps,
		Scenarios:     scenarios,
		ToolMentions:  mentions,
	}
}

func splitSections(body string) map[string]string {
	out := map[string]string{}
	idx := sectionRe.FindAllStringSubmatchIndex(body, -1)
	for i, loc := range idx {
		if len(loc) < 4 {
			continue
		}
		title := strings.ToLower(strings.TrimSpace(body[loc[2]:loc[3]]))
		start := loc[1]
		end := len(body)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		out[title] = strings.TrimSpace(body[start:end])
	}
	return out
}

func extractBullets(block string) []string {
	var out []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			out = append(out, strings.TrimSpace(line[2:]))
		}
	}
	return out
}

func extractSteps(workflow string) []domain.ProcedureStep {
	var steps []domain.ProcedureStep
	parts := stepRe.Split(workflow, -1)
	if len(parts) <= 1 {
		return steps
	}
	for i := 1; i+2 <= len(parts); i += 3 {
		num, _ := strconv.Atoi(strings.TrimSpace(parts[i]))
		title := strings.TrimSpace(parts[i+1])
		content := strings.TrimSpace(parts[i+2])
		mentions := tokenizeTools(content)
		kind := domain.StepManual
		if strings.Contains(content, "```bash") || strings.Contains(content, "```sh") {
			kind = domain.StepShell
		}
		if len(mentions) > 0 {
			kind = domain.StepTool
		}
		if len(content) > 2000 {
			content = content[:2000]
		}
		steps = append(steps, domain.ProcedureStep{
			Number:       num,
			Title:        title,
			Kind:         kind,
			Body:         content,
			ToolMentions: mentions,
		})
	}
	return steps
}

func tokenizeTools(text string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range toolTokenRe.FindAllString(text, -1) {
		t := strings.ToLower(m)
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func collectMentions(steps []domain.ProcedureStep, extra string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range steps {
		for _, t := range s.ToolMentions {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	for _, t := range tokenizeTools(extra) {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
