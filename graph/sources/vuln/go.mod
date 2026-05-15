module github.com/butbeautifulv/threat_intelligence/graph/sources/vuln

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1 v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/neo4jclient v0.0.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
)

replace github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1 => ../../contract/ingestv1

replace github.com/butbeautifulv/threat_intelligence/graph/neo4jclient => ../../neo4jclient
