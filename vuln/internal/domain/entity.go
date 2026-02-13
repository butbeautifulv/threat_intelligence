package domain

type Vulnerability struct {
	ID       string
	CVE      string
	Summary  string
	CWE      []string
	CPEs     []CPE
	CVSS     *CVSS
	Exploits []ExploitRef
	//Sources  []SourceRef
}

type CVSS struct {
	Version string
	Base    float64
	Vector  string
}

type CPE struct {
	URI string
}

type ExploitRef struct {
	Source string // metasploit, exploitdb, vulners
	RefID  string
	URL    string
}
