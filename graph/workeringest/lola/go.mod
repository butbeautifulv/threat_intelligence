module github.com/butbeautifulv/threat_intelligence/graph/workeringest/lola

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1 v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/sources/lola v0.0.0
)

require github.com/neo4j/neo4j-go-driver/v5 v5.28.4 // indirect

replace github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1 => ../../contract/ingestv1

replace github.com/butbeautifulv/threat_intelligence/graph/sources/lola => ../../sources/lola
