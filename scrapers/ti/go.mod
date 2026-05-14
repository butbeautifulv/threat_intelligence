module ti

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/graph v0.0.0
	github.com/butbeautifulv/threat_intelligence/pkg/ingestv1 v0.0.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	ingestpub v0.0.0
)

require (
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/nats-io/nats.go v1.39.1 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/graph => ../../graph

replace github.com/butbeautifulv/threat_intelligence/pkg/ingestv1 => ../../pkg/ingestv1

replace ingestpub => ../ingestpub
