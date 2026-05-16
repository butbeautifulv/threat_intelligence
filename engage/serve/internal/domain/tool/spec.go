package tool

import "github.com/butbeautifulv/veil/pkg/engage/toolid"

// Param describes a tool input field (catalog + MCP schema).
type Param struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Default  string `yaml:"default,omitempty"`
	Required bool   `yaml:"required"`
}

// Spec describes a catalog tool entry (no I/O).
type Spec struct {
	Name         string
	Category     toolid.Category
	Binary       string
	ArgsTemplate []string
	Parameters   []Param
	TimeoutSec   int
	Description  string
	Enabled      bool
}

// InputSchema builds MCP JSON Schema from parameters.
func (s Spec) InputSchema() map[string]any {
	props := map[string]any{}
	var required []string
	for _, p := range s.Parameters {
		t := p.Type
		if t == "" {
			t = "string"
		}
		prop := map[string]any{"type": t, "description": p.Name}
		if p.Default != "" {
			prop["default"] = p.Default
		}
		props[p.Name] = prop
		if p.Required {
			required = append(required, p.Name)
		}
	}
	if len(props) == 0 {
		props["target"] = map[string]any{"type": "string", "description": "Target host, URL, or domain"}
		props["additional_args"] = map[string]any{"type": "string", "description": "Extra CLI flags"}
		required = []string{"target"}
	}
	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// DefaultParameters returns parameter values from catalog defaults.
func (s Spec) DefaultParameters() map[string]string {
	out := make(map[string]string, len(s.Parameters))
	for _, p := range s.Parameters {
		if p.Default != "" {
			out[p.Name] = p.Default
		}
	}
	return out
}
