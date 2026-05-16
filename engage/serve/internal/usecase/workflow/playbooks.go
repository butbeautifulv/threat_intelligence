package workflow

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Playbook describes a named bug bounty workflow template.
type Playbook struct {
	Name        string `yaml:"name"`
	Objective   string `yaml:"objective"`
	Workflow    string `yaml:"workflow"`
	MaxTools    int    `yaml:"max_tools"`
	Description string `yaml:"description"`
}

type playbookFile struct {
	Playbooks []Playbook `yaml:"playbooks"`
}

// LoadPlaybooks reads YAML playbooks from path (default engage/serve/playbooks/bugbounty.yaml).
func LoadPlaybooks(path string) ([]Playbook, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var f playbookFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return f.Playbooks, nil
}

// DefaultPlaybooksPath resolves playbooks next to catalog when ENGAGE_PLAYBOOKS_PATH unset.
func DefaultPlaybooksPath(catalogPath string) string {
	if v := os.Getenv("ENGAGE_PLAYBOOKS_PATH"); v != "" {
		return v
	}
	dir := filepath.Dir(catalogPath)
	return filepath.Join(dir, "..", "playbooks", "bugbounty.yaml")
}

// LoadAllPlaybooks merges bug bounty and CTF playbook YAML files.
func LoadAllPlaybooks(catalogPath string) ([]Playbook, error) {
	dir := filepath.Dir(DefaultPlaybooksPath(catalogPath))
	var all []Playbook
	for _, name := range []string{"bugbounty.yaml", "ctf.yaml"} {
		list, err := LoadPlaybooks(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		all = append(all, list...)
	}
	if v := os.Getenv("ENGAGE_PLAYBOOKS_PATH"); v != "" {
		list, err := LoadPlaybooks(v)
		if err != nil {
			return nil, err
		}
		all = append(all, list...)
	}
	return all, nil
}

// FindPlaybook returns a playbook by name.
func FindPlaybook(list []Playbook, name string) (Playbook, bool) {
	for _, p := range list {
		if p.Name == name || p.Workflow == name {
			return p, true
		}
	}
	return Playbook{}, false
}
