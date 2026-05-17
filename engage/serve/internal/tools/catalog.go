package tools

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/butbeautifulv/veil/pkg/engage/domain/tool"
	"github.com/butbeautifulv/veil/pkg/engage/toolid"
)

type catalogFile struct {
	Tools []catalogEntry `yaml:"tools"`
}

type catalogEntry struct {
	Name        string        `yaml:"name"`
	Category    string        `yaml:"category"`
	Binary      string        `yaml:"binary"`
	Args        []string      `yaml:"args"`
	Parameters  []tool.Param  `yaml:"parameters"`
	TimeoutSec  int           `yaml:"timeout_sec"`
	Description string        `yaml:"description"`
	Enabled     *bool         `yaml:"enabled"`
}

// LoadCatalog reads one or more YAML catalog files (later files override same name).
func LoadCatalog(paths ...string) ([]tool.Spec, error) {
	byName := make(map[string]tool.Spec)
	for _, path := range paths {
		specs, err := loadCatalogFile(path)
		if err != nil {
			return nil, err
		}
		for _, s := range specs {
			byName[s.Name] = s
		}
	}
	out := make([]tool.Spec, 0, len(byName))
	for _, s := range byName {
		out = append(out, s)
	}
	return out, nil
}

func loadCatalogFile(path string) ([]tool.Spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cf catalogFile
	if err := yaml.Unmarshal(b, &cf); err != nil {
		return nil, err
	}
	out := make([]tool.Spec, 0, len(cf.Tools))
	for _, e := range cf.Tools {
		enabled := true
		if e.Enabled != nil {
			enabled = *e.Enabled
		}
		cat := toolid.Category(e.Category)
		if !cat.Valid() {
			return nil, fmt.Errorf("tool %q: invalid category %q", e.Name, e.Category)
		}
		timeout := e.TimeoutSec
		if timeout <= 0 {
			timeout = 300
		}
		params := e.Parameters
		if len(params) == 0 {
			params = defaultParamsForTool(e.Name)
		}
		args := e.Args
		if len(args) == 0 {
			args = defaultArgsForTool(e.Name)
		}
		out = append(out, tool.Spec{
			Name:         e.Name,
			Category:     cat,
			Binary:       e.Binary,
			ArgsTemplate: args,
			Parameters:   params,
			TimeoutSec:   timeout,
			Description:  e.Description,
			Enabled:      enabled,
		})
	}
	return out, nil
}

func defaultParamsForTool(name string) []tool.Param {
	base := []tool.Param{
		{Name: "target", Type: "string", Required: true},
		{Name: "additional_args", Type: "string"},
	}
	switch name {
	case "nmap_scan":
		return []tool.Param{
			{Name: "target", Type: "string", Required: true},
			{Name: "scan_type", Type: "string", Default: "-sV"},
			{Name: "ports", Type: "string"},
			{Name: "additional_args", Type: "string", Default: "-T4 -Pn"},
		}
	case "nuclei_scan":
		return append(base, tool.Param{Name: "templates", Type: "string"})
	default:
		return base
	}
}

func defaultArgsForTool(name string) []string {
	switch name {
	case "nmap_scan":
		return []string{"{scan_type}", "-p", "{ports}", "{additional_args}", "{target}"}
	case "nuclei_scan":
		return []string{"-u", "{target}", "-t", "{templates}", "{additional_args}"}
	default:
		return []string{"{target}", "{additional_args}"}
	}
}
