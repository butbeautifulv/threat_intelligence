package tiingest

import "github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tidomain"

// Envelope is a single JSONL record (ioc, campaign, cluster, actor, or report).
type Envelope struct {
	IOC      *tidomain.IOC      `json:"ioc,omitempty"`
	Campaign *tidomain.Campaign `json:"campaign,omitempty"`
	Cluster  *tidomain.Cluster  `json:"cluster,omitempty"`
	Actor    *tidomain.Actor    `json:"actor,omitempty"`
	Report   *tidomain.Report   `json:"report,omitempty"`
}
