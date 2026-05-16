package target

// Target is a scan subject (host, URL, CIDR, etc.).
type Target struct {
	Value string `json:"value"`
	Kind  string `json:"kind,omitempty"`
}
