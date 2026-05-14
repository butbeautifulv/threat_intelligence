module nuclei

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/graph v0.0.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/graph => ../../graph
