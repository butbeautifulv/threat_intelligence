module github.com/butbeautifulv/veil/pipeline/ned

go 1.25.0

require (
	github.com/butbeautifulv/veil/pipeline/connector v0.0.0
	github.com/butbeautifulv/veil/pipeline/pkg v0.0.0
	github.com/butbeautifulv/veil/pkg v0.0.0
	github.com/nats-io/nats.go v1.48.0
	golang.org/x/sync v0.20.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace github.com/butbeautifulv/veil/pipeline/connector => ../connector

replace github.com/butbeautifulv/veil/pipeline/pkg => ../pkg

replace github.com/butbeautifulv/veil/pkg => ../../pkg
