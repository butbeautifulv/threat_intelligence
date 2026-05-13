module lola

go 1.25.0

replace github.com/butbeautifulv/threat_intelligence/graph => ../../graph

require (
	github.com/butbeautifulv/threat_intelligence/graph v0.0.0
	github.com/fatih/color v1.18.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	golang.org/x/exp v0.0.0-20260212183809-81e46e3db34a
	golang.org/x/net v0.50.0
	golang.org/x/sync v0.19.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.41.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)
