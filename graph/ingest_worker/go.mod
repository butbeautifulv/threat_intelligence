module github.com/butbeautifulv/threat_intelligence/graph/ingest_worker

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1 v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/internal/natsensure v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/storage/coderules v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/storage/nuclei v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/storage/sbom v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/workeringest/ds v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/workeringest/lola v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/workeringest/ti v0.0.0
	github.com/butbeautifulv/threat_intelligence/graph/workeringest/vuln v0.0.0
	github.com/nats-io/nats.go v1.39.1
	golang.org/x/sync v0.20.0
)

require (
	github.com/butbeautifulv/threat_intelligence/graph/sources/ds v0.0.0 // indirect
	github.com/butbeautifulv/threat_intelligence/graph/sources/lola v0.0.0 // indirect
	github.com/butbeautifulv/threat_intelligence/graph/sources/ti v0.0.0 // indirect
	github.com/butbeautifulv/threat_intelligence/graph/sources/vuln v0.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/graph/contract/ingestv1 => ../contract/ingestv1

replace github.com/butbeautifulv/threat_intelligence/graph/internal/natsensure => ../internal/natsensure

replace github.com/butbeautifulv/threat_intelligence/graph/sources/ds => ../sources/ds

replace github.com/butbeautifulv/threat_intelligence/graph/sources/lola => ../sources/lola

replace github.com/butbeautifulv/threat_intelligence/graph/sources/ti => ../sources/ti

replace github.com/butbeautifulv/threat_intelligence/graph/sources/vuln => ../sources/vuln

replace github.com/butbeautifulv/threat_intelligence/graph/storage/coderules => ../storage/coderules

replace github.com/butbeautifulv/threat_intelligence/graph/storage/nuclei => ../storage/nuclei

replace github.com/butbeautifulv/threat_intelligence/graph/storage/sbom => ../storage/sbom

replace github.com/butbeautifulv/threat_intelligence/graph/workeringest/ds => ../workeringest/ds

replace github.com/butbeautifulv/threat_intelligence/graph/workeringest/lola => ../workeringest/lola

replace github.com/butbeautifulv/threat_intelligence/graph/workeringest/ti => ../workeringest/ti

replace github.com/butbeautifulv/threat_intelligence/graph/workeringest/vuln => ../workeringest/vuln
