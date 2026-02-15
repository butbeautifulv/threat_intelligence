vuln: {
    Vulnerability: {
        id:      string
        cve:     string
        summary: string

        cwe?: [...string]

        cpes?: [...{
            uri: string
        }]

        cvss?: {
            version: string
            base:    number
            vector:  string
        }

        exploits?: [...{
            source: string // metasploit, exploitdb, vulners
            refID:  string
            url:    string
        }]
    }
}
