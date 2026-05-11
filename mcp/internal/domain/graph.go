package domain

type Node struct {
	ElementID string                 `json:"id"`
	Labels    []string               `json:"labels"`
	Props     map[string]any         `json:"props"`
	Kind      string                 `json:"kind,omitempty"`
	Title     string                 `json:"title,omitempty"`
	Markdown  string                 `json:"markdown,omitempty"`
}

type Edge struct {
	ElementID string `json:"id"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Target    string `json:"target"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

