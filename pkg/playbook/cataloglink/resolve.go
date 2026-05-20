package cataloglink

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

var (
	catalogOnce sync.Once
	catalogNames map[string]struct{}
	aliases      = map[string]string{
		"nmap":          "nmap_scan",
		"nuclei":        "nuclei_scan",
		"nikto":         "nikto_scan",
		"sqlmap":        "sqlmap_scan",
		"gobuster":      "gobuster_scan",
		"ffuf":          "ffuf_scan",
		"feroxbuster":   "feroxbuster_scan",
		"wpscan":        "wpscan_scan",
		"masscan":       "masscan_scan",
		"httpx":         "httpx_probe",
		"subfinder":     "subfinder_scan",
		"amass":         "amass_enum",
		"theharvester":  "theharvester_osint",
	}
	toolNameLineRe = regexp.MustCompile(`^  - name:\s+(\S+)`)
	toolBinaryRe   = regexp.MustCompile(`^    binary:\s+(\S+)`)
)

var catalogRepoRoot = pbindex.RepoRoot

func loadCatalog() {
	root, err := catalogRepoRoot()
	if err != nil {
		catalogNames = map[string]struct{}{}
		return
	}
	catalogNames = map[string]struct{}{}
	for _, rel := range []string{
		"engage/serve/catalog/tools.yaml",
		"engage/serve/catalog/tools.live.yaml",
	} {
		path := filepath.Join(root, rel)
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			if m := toolNameLineRe.FindStringSubmatch(line); len(m) > 1 {
				catalogNames[m[1]] = struct{}{}
			}
			if m := toolBinaryRe.FindStringSubmatch(line); len(m) > 1 {
				catalogNames[m[1]] = struct{}{}
			}
		}
	}
}

// ResetCatalogForTest clears the lazy catalog cache (tests only).
func ResetCatalogForTest() {
	catalogOnce = sync.Once{}
	catalogNames = nil
}

// ResolveMentions maps tool tokens to engage catalog tool names (read-only).
func ResolveMentions(mentions []string) []string {
	catalogOnce.Do(loadCatalog)
	seen := map[string]struct{}{}
	var out []string
	for _, m := range mentions {
		t := strings.ToLower(strings.TrimSpace(m))
		if t == "" {
			continue
		}
		hit := resolveOne(t)
		if hit == "" {
			continue
		}
		if _, dup := seen[hit]; dup {
			continue
		}
		seen[hit] = struct{}{}
		out = append(out, hit)
	}
	return out
}

func resolveOne(token string) string {
	if hit, ok := aliases[token]; ok {
		if _, ok := catalogNames[hit]; ok {
			return hit
		}
	}
	if len(token) < 3 {
		return ""
	}
	names := make([]string, 0, len(catalogNames))
	for name := range catalogNames {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		low := strings.ToLower(name)
		if low == token || strings.HasSuffix(low, "_"+token) || strings.HasPrefix(low, token+"_") {
			return name
		}
	}
	return ""
}
