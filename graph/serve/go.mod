module github.com/butbeautifulv/threat_intelligence/graph/serve

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/graph/connector v0.0.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
)

replace github.com/butbeautifulv/threat_intelligence/graph/connector => ../connector
