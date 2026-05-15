# nvdparse

Shared parser for **NVD CVE API 2.0** JSON pages (`vulnerabilities[]`).

Used by:

- [scrape/sources/vuln/internal/usecase/scrape.go](../../scrape/sources/vuln/internal/usecase/scrape.go) — page stats / logging
- [pipeline/pipeline_worker/internal/handle/vuln.go](../../pipeline/pipeline_worker/internal/handle/vuln.go) — `KindVulnNVDPage` → full `ingestv1` upserts

Extracts per CVE: `Summary`, `CWE[]`, `CPEs[]`, optional `CVSS` (v3.1 / v3.0).

```go
vulns, totalResults, err := nvdparse.ParsePage(rawJSON)
```

Tests: `go test ./...` with [testdata/nvd_page_min.json](testdata/nvd_page_min.json).
