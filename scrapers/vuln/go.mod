module vuln

go 1.25.0

replace github.com/butbeautifulv/threat_intelligence/graph => ../../graph

replace github.com/butbeautifulv/threat_intelligence/pkg/ingestv1 => ../../pkg/ingestv1

replace ingestpub => ../ingestpub

require (
	github.com/butbeautifulv/threat_intelligence/graph v0.0.0
	github.com/butbeautifulv/threat_intelligence/pkg/ingestv1 v0.0.0
	github.com/fatih/color v1.18.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	go.mongodb.org/mongo-driver v1.11.4
	golang.org/x/exp v0.0.0-20260212183809-81e46e3db34a
	golang.org/x/sync v0.19.0
	ingestpub v0.0.0
)

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/nats-io/nats.go v1.39.1 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
