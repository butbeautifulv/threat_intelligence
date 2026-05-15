package domain

// Actor is a first-class threat actor (group) with stable id when omitted (derived from name).
type Actor struct {
	ID          string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string   `json:"name" yaml:"name"`
	Aliases     []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Source      string   `json:"source,omitempty" yaml:"source,omitempty"`
}

// Report is a public write-up (blog, vendor research) used for provenance.
type Report struct {
	ID           string `json:"id,omitempty" yaml:"id,omitempty"`
	Title        string `json:"title" yaml:"title"`
	Provider     string `json:"provider" yaml:"provider"`
	Link         string `json:"link" yaml:"link"`
	PublishedAt  string `json:"published_at,omitempty" yaml:"published_at,omitempty"`
	BodyMarkdown string `json:"body_markdown,omitempty" yaml:"body_markdown,omitempty"`
	Source       string `json:"source,omitempty" yaml:"source,omitempty"`
}
