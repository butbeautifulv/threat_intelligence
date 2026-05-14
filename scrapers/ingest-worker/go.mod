module ingest-worker

go 1.25.0

require (
	coderules v0.0.0
	ds v0.0.0
	github.com/butbeautifulv/threat_intelligence/pkg/ingestv1 v0.0.0
	github.com/nats-io/nats.go v1.39.1
	golang.org/x/sync v0.20.0
	ingestpub v0.0.0
	lola v0.0.0
	nuclei v0.0.0
	sbom v0.0.0
	ti v0.0.0
	vuln v0.0.0
)

require (
	github.com/butbeautifulv/threat_intelligence/graph v0.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace coderules => ../coderules

replace ds => ../ds

replace github.com/butbeautifulv/threat_intelligence/graph => ../../graph

replace github.com/butbeautifulv/threat_intelligence/pkg/ingestv1 => ../../pkg/ingestv1

replace ingestpub => ../ingestpub

replace lola => ../lola

replace nuclei => ../nuclei

replace sbom => ../sbom

replace ti => ../ti

replace vuln => ../vuln
