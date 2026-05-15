module github.com/butbeautifulv/veil/graph/serve

go 1.25.0

require (
	github.com/butbeautifulv/veil/graph/connector v0.0.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
)

replace github.com/butbeautifulv/veil/graph/connector => ../connector
