module github.com/butbeautifulv/threat_intelligence/scrape/scrape_worker

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/scrape/factory v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/coderules v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/ds v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/lola v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/nuclei v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/sbom v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/ti v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln v0.0.0
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/scrape/factory => ../factory

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/coderules => ../sources/coderules

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/ds => ../sources/ds

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/lola => ../sources/lola

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/nuclei => ../sources/nuclei

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/sbom => ../sources/sbom

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/ti => ../sources/ti

replace github.com/butbeautifulv/threat_intelligence/scrape/sources/vuln => ../sources/vuln
