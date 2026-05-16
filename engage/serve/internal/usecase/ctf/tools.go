package ctf

import (
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
)

// ToolManager suggests CTF tools by category and description keywords.
type ToolManager struct{}

func NewToolManager() *ToolManager {
	return &ToolManager{}
}

// categoryTools maps CTF categories to tool id groups (short ids, resolved via catalog).
var categoryTools = map[string]map[string][]string{
	"web": {
		"reconnaissance":         {"httpx", "katana", "gau"},
		"vulnerability_scanning": {"nuclei", "dalfox", "sqlmap", "nikto"},
		"content_discovery":      {"gobuster", "feroxbuster"},
		"parameter_testing":      {"arjun", "paramspider"},
	},
	"crypto": {
		"hash_analysis":      {"hashcat", "john"},
		"cipher_analysis":    {"hash-identifier"},
		"encoding":           {"base64", "rot13"},
	},
	"pwn": {
		"binary_analysis":    {"checksec", "strings", "file"},
		"exploit_development": {"pwntools", "ropper"},
		"static_analysis":    {"ghidra", "radare2"},
	},
	"forensics": {
		"file_analysis":    {"binwalk", "file", "strings"},
		"image_forensics":  {"exiftool", "steghide", "zsteg"},
		"metadata":         {"exiftool"},
	},
	"rev": {
		"disassemblers": {"ghidra", "radare2"},
		"analysis":     {"strings", "checksec"},
	},
	"misc": {
		"encoding": {"base64", "rot13"},
	},
	"osint": {
		"search_engines": {"theharvester", "amass", "subfinder"},
	},
}

// toolCommands are template command hints (not executed directly).
var toolCommands = map[string]string{
	"httpx":     "httpx -probe -tech-detect -status-code -title",
	"nuclei":    "nuclei -severity critical,high",
	"sqlmap":    "sqlmap --batch",
	"checksec":  "checksec --file",
	"strings":   "strings -n 8",
	"binwalk":   "binwalk -e",
	"hashcat":   "hashcat -m 0 -a 0",
	"gobuster":  "gobuster dir",
	"exiftool":  "exiftool -all",
}

// SuggestTools returns tool ids for a challenge description and category.
func (m *ToolManager) SuggestTools(description, category string) []string {
	category = NormalizeCategory(category)
	if category == "" {
		category = "misc"
	}
	desc := strings.ToLower(description)
	seen := map[string]struct{}{}
	var out []string
	add := func(ids ...string) {
		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}

	for _, ids := range categoryTools[category] {
		add(ids...)
	}

	switch category {
	case "web":
		if strings.Contains(desc, "sql") || strings.Contains(desc, "injection") {
			add("sqlmap")
		}
		if strings.Contains(desc, "xss") {
			add("dalfox")
		}
		if strings.Contains(desc, "wordpress") || strings.Contains(desc, "wp") {
			add("wpscan")
		}
		if strings.Contains(desc, "directory") || strings.Contains(desc, "hidden") {
			add("gobuster", "feroxbuster")
		}
	case "crypto":
		if strings.Contains(desc, "hash") || strings.Contains(desc, "md5") || strings.Contains(desc, "sha") {
			add("hashcat", "john")
		}
		if strings.Contains(desc, "rsa") {
			add("rsatool")
		}
		if strings.Contains(desc, "base64") {
			add("base64")
		}
		if strings.Contains(desc, "rot") || strings.Contains(desc, "caesar") {
			add("rot13")
		}
	case "pwn":
		add("checksec", "strings", "file")
		if strings.Contains(desc, "buffer") || strings.Contains(desc, "overflow") {
			add("pwntools", "ropper")
		}
		if strings.Contains(desc, "format") {
			add("pwntools")
		}
	case "forensics":
		add("binwalk", "exiftool", "strings")
		if strings.Contains(desc, "stego") || strings.Contains(desc, "image") {
			add("steghide", "zsteg")
		}
	}

	return out
}

// ResolveTools maps short ids to catalog names present in registry.
func (m *ToolManager) ResolveTools(ids []string, reg *tools.Registry) []string {
	return tools.ResolveCatalogNames(ids, reg)
}

// ToolCommand returns a template command string for a tool id.
func (m *ToolManager) ToolCommand(tool, target string) string {
	if tpl, ok := toolCommands[tool]; ok {
		if target != "" {
			return tpl + " " + target
		}
		return tpl
	}
	if target != "" {
		return tool + " " + target
	}
	return tool
}

// CategoryToolsFlat returns all tool ids for a category group key prefix.
func (m *ToolManager) CategoryToolsFlat(category string) []string {
	category = NormalizeCategory(category)
	seen := map[string]struct{}{}
	var out []string
	groups, ok := categoryTools[category]
	if !ok {
		return nil
	}
	for _, ids := range groups {
		for _, id := range ids {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}
