module github.com/butbeautifulv/threat_intelligence/scrape/feeds

go 1.25.0

require github.com/butbeautifulv/threat_intelligence/scrape/ledger v0.0.0

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.0 // indirect
)

replace github.com/butbeautifulv/threat_intelligence/scrape/ledger => ../ledger
