module github.com/butbeautifulv/threat_intelligence/pipeline/pipeline_worker

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/pkg/nvdparse v0.0.0
	github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1 v0.0.0
	github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/appsec v0.0.0
	github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/ti v0.0.0
	github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tidomain v0.0.0
	github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tiingest v0.0.0
	github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/vuln v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1 v0.0.0
	github.com/nats-io/nats.go v1.39.1
	golang.org/x/sync v0.20.0
	gopkg.in/yaml.v3 v3.0.1
	ingestpub v0.0.0
)

require (
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/pkg/nvdparse => ../../pkg/nvdparse

replace github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1 => ../contract/ingestv1

replace github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1 => ../../scrape/contract/scrapev1

replace github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/appsec => ../internal/normalize/appsec

replace github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/ti => ../internal/normalize/ti

replace github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tidomain => ../internal/normalize/tidomain

replace github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tiingest => ../internal/normalize/tiingest

replace github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/vuln => ../internal/normalize/vuln

replace ingestpub => ../pub
