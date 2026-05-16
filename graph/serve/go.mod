module github.com/butbeautifulv/veil/graph/serve

go 1.25.0

require (
	github.com/MicahParks/keyfunc/v3 v3.8.0
	github.com/butbeautifulv/veil/graph/connector v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	golang.org/x/time v0.9.0 // indirect
)

replace github.com/butbeautifulv/veil/graph/connector => ../connector
