module github.com/butbeautifulv/threat_intelligence/scrape/sources/coderules

go 1.25.0

require (
	github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1 v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/factory v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/feeds v0.0.0
	github.com/butbeautifulv/threat_intelligence/scrape/ledger v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/nats-io/nats.go v1.39.1 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	scrapepub v0.0.0 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1 => ../../contract/scrapev1

replace github.com/butbeautifulv/threat_intelligence/scrape/factory => ../../factory

replace github.com/butbeautifulv/threat_intelligence/scrape/feeds => ../../feeds

replace github.com/butbeautifulv/threat_intelligence/scrape/ledger => ../../ledger

replace scrapepub => ../../pub
