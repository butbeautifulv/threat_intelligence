package domain

type IOCType string

const (
	IOCIP     IOCType = "ip"
	IOCDomain IOCType = "domain"
	IOCURL    IOCType = "url"
	IOCHash   IOCType = "hash"
)

type IOC struct {
	Type       IOCType  `json:"type" yaml:"type"`
	Value      string   `json:"value" yaml:"value"`
	Confidence *int     `json:"confidence,omitempty" yaml:"confidence,omitempty"`
	Tags       []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Source     string   `json:"source,omitempty" yaml:"source,omitempty"`
}

type Campaign struct {
	ID      string   `json:"id" yaml:"id"`
	Name    string   `json:"name" yaml:"name"`
	Actors  []string `json:"actors,omitempty" yaml:"actors,omitempty"`
	IOCs    []IOC    `json:"iocs,omitempty" yaml:"iocs,omitempty"`
	Summary string   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Source  string   `json:"source,omitempty" yaml:"source,omitempty"`
}

type Cluster struct {
	ID          string     `json:"id" yaml:"id"`
	Name        string     `json:"name" yaml:"name"`
	Campaigns   []Campaign `json:"campaigns,omitempty" yaml:"campaigns,omitempty"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Source      string     `json:"source,omitempty" yaml:"source,omitempty"`
}

