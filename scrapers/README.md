# Scrapers (ingest → Neo4j)

Go services that pull public data and write into the shared Neo4j database. Built via [docker/](../docker/) Dockerfiles with repo root as context.

## Layout

| Directory | Binary | Role |
|-----------|--------|------|
| [vuln/](vuln/) | `vuln` | NVD CVE 2.0, Metasploit paths, Exploit-DB CSV, optional Vulners |
| [lola/](lola/) | `lola` | LOLBAS, GTFOBins, LOFTS, MITRE ATT&CK STIX enterprise |
| [ds/](ds/) | `ds` | SigmaHQ, YARA (Neo23x0), Atomic Red Team, Caldera Stockpile abilities |
| [ti/](ti/) | `ti` | CISA KEV, URLhaus, PT RSS, ThreatFox, MalwareBazaar, Feodo, OpenPhish, optional JSONL file |
| [proxybroker/](proxybroker/) | `proxybroker` | HTTP proxy pool for scrapers (Compose service name `proxybroker`) |
| [cue_schemas/](cue_schemas/) | — | Cue schemas (`merge.cue` imports `schema/ds.cue`) |

## Sources matrix

| Source | Service | Parser status | Public endpoint / repo (reference) | Notes |
|--------|---------|----------------|--------------------------------------|-------|
| NVD CVE API 2.0 | `vuln` | Implemented | `https://services.nvd.nist.gov/rest/json/cves/2.0` | `NVD_API_KEY`, `NVD_MAX_PAGES`, `VULN_REQUEST_DELAY`, `VULN_CACHE_DIR` |
| Metasploit modules (GitHub API) | `vuln` | Implemented | `https://api.github.com/repos/rapid7/metasploit-framework/contents/...` | `GITHUB_TOKEN`, `VULN_METASPLOIT_MAX_RB`, `VULN_METASPLOIT_MAX_DIRS` |
| Exploit-DB CSV (GitLab mirror) | `vuln` | Implemented | Default CSV: `https://gitlab.com/exploit-database/exploitdb/-/raw/main/files_exploits.csv` | `VULN_EXPLOITDB_CSV_URL`, `VULN_EXPLOITDB_MAX_ROWS` |
| Exploit-DB web (links in graph) | `vuln` | Implemented (URLs only) | `https://www.exploit-db.com/exploits/<id>` | Stored as metadata on ingested rows |
| Vulners search API | `vuln` | Implemented when key set | `https://vulners.com/api/v3/search/lucene/` | `VULNERS_API_KEY`; `VULN_VULNERS_MAX`, `VULN_VULNERS_QUERY` |
| LOLBAS (GitHub API) | `lola` | Implemented | `https://api.github.com/repos/LOLBAS-Project/LOLBAS/contents/...` | `LOLA_CACHE_DIR`, `GITHUB_TOKEN` |
| GTFOBins (GitHub API) | `lola` | Implemented | `https://api.github.com/repos/GTFOBins/GTFOBins.github.io/contents/...` | Same as LOLBAS |
| LOFTS | `lola` | Implemented | Default: `https://lofts.galeal.com/` | `LOFTS_URL`, `LOFTS_SKIP_ON_ERROR`, `LOFTS_MAX_ENTRIES` |
| MITRE ATT&CK (STIX bundle) | `lola` | Implemented | Default: `https://raw.githubusercontent.com/mitre-attack/attack-stix-data/master/enterprise-attack/enterprise-attack.json` | `LOLA_MITRE_STIX_URL`, `LOLA_MITRE_MAX_TECHNIQUES`, `LOLA_MITRE_MAX_RELATIONSHIPS` |
| MITRE ATT&CK (human-readable) | — | Reference only | `https://attack.mitre.org/techniques/` | Used in docs / manual correlation; STIX URL above is what `lola` ingests |
| SigmaHQ (GitHub API) | `ds` | Implemented | `https://api.github.com/repos/SigmaHQ/sigma/contents/rules/windows/process_creation` | `DS_MAX_SIGMA`, `DS_CACHE_DIR`, `GITHUB_TOKEN` |
| Neo23x0 signature-base YARA (GitHub API) | `ds` | Implemented | `https://api.github.com/repos/Neo23x0/signature-base/contents/yara` (fallback `iocs/yara`) | `DS_MAX_YARA`, `DS_YARA_PATH`, `GITHUB_TOKEN` |
| Atomic Red Team (GitHub API) | `ds` | Implemented | `https://api.github.com/repos/redcanaryco/atomic-red-team/contents/atomics` | `DS_MAX_ATOMIC`, `DS_CACHE_DIR`, `GITHUB_TOKEN` |
| Caldera Stockpile abilities (GitHub API) | `ds` | Implemented | `https://api.github.com/repos/mitre/caldera/contents/plugins/stockpile/data/abilities` | `DS_MAX_CALDERA`, `DS_CALDERA_BASE_PATH`, `GITHUB_TOKEN` |
| CISA KEV JSON | `ti` | Implemented | `https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json` | `TI_KEV_MAX`, `TI_CACHE_DIR`, `TI_REQUEST_DELAY` |
| URLhaus recent CSV | `ti` | Implemented | `https://urlhaus.abuse.ch/downloads/csv_recent/` | `TI_URLHAUS_MAX`; full DB dump (not used by parser): `https://urlhaus.abuse.ch/downloads/csv/` |
| Positive Technologies RSS | `ti` | Implemented | Override via `PT_RSS_URL`; default in code: `https://www.ptsecurity.com/rss/all.xml` | `TI_PT_MAX`, `TI_FEEDS` |
| ThreatFox | `ti` | Implemented | Public export: `https://threatfox.abuse.ch/export/json/recent/`; or API `https://threatfox-api.abuse.ch/api/v1/` when `THREATFOX_AUTH_KEY` set | `TI_THREATFOX_MAX`, `TI_THREATFOX_DAYS` (1–7, API only), file cache under `TI_CACHE_DIR` |
| MalwareBazaar recent | `ti` | Implemented when key set | POST `https://mb-api.abuse.ch/api/v1/` (`query=get_recent`) | **Required:** `MALWAREBAZAAR_AUTH_KEY` or `MALWARE_BAZAAR_API_KEY` (abuse.ch Auth-Key header). `TI_MALWAREBAZAAR_MAX` |
| Feodo Tracker IP blocklist | `ti` | Implemented | Default: `https://feodotracker.abuse.ch/downloads/ipblocklist_recommended.txt` | `TI_FEODO_MAX`, optional `FEODO_BLOCKLIST_URL` |
| OpenPhish URL feed | `ti` | Implemented | Default: `https://openphish.com/feed.txt` | `TI_OPENPHISH_MAX`, optional `OPENPHISH_FEED_URL` |
| TI JSONL (local / mounted file) | `ti` | Implemented | — | `--input path.jsonl` (compose mounts [example.jsonl](ti/example.jsonl) as `/app/example.jsonl`) |

**TI feeds vs Docker:** root [docker-compose.yml](../docker-compose.yml) defaults `TI_FEEDS=kev,urlhaus,threatfox,malwarebazaar,feodo`. Append `openphish` when you want phishing URLs (the remote feed can be slow or flaky; the scraper logs a warning and continues if the download fails). MalwareBazaar is skipped unless `MALWAREBAZAAR_AUTH_KEY` (or `MALWARE_BAZAAR_API_KEY`) is set. With `THREATFOX_AUTH_KEY`, ThreatFox uses the authenticated API instead of the public JSON export. Use `TI_PROXY_URLS` (and optional `TI_PROXY_MODE=only`) to route traffic via [proxybroker](proxybroker/). If the default PT URL returns HTML errors, add `pt` explicitly to `TI_FEEDS` only when needed, or set `PT_RSS_URL` to a stable RSS endpoint you operate.

### Optional: same stack via Docker Compose

From repo root, `docker compose up --build` runs **Neo4j** + **graph pack import** + **HTTP API** by default (no live scraping). See [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md).

Re-run **scrapers** and **`proxybroker`** (all require `--profile scrape`; they never start on default `docker compose up`):

```bash
docker compose --profile scrape up --build -d
```

Re-run ingest only: `docker compose restart vuln lola ds ti` (with profile `scrape` enabled for those services).

### Planned only (no parser here)

- AlienVault OTX API — future.
- MISP feeds — future.

## Run locally (Neo4j must be up)

From repo root (paths under `scrapers/`):

```bash
cd scrapers/proxybroker && go run ./cmd
cd scrapers/vuln && go run ./cmd
cd scrapers/lola && go run ./cmd
cd scrapers/ds && go run ./cmd
cd scrapers/ti && go run ./cmd --feeds kev,urlhaus,threatfox,feodo,openphish --input example.jsonl
```

`proxybroker` does not talk to Neo4j; start it when scrapers need `*_PROXY_URLS`.

`go.work` at repo root includes these modules; use `go work sync` if imports drift.

## Graph export and packs

**Consuming a pack** (no scrapers): use the default Compose stack — `graph-bootstrap` imports a ZIP before `api` starts. Env vars and bind mounts: [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md).

**Building a pack** (scraping): enable the Compose **`scrape`** profile (`proxybroker` + ingest services), fill Neo4j, then run the scripts below on the host.

Full export/pack/import flow lives in root [scripts/](../scripts/) (see root [README.md](../README.md)):

- `../scripts/export-graph-cypher.sh` — needs running Neo4j from compose.
- `../scripts/build-graph-pack.sh` / `../scripts/import-graph-pack.sh` — versioned ZIP + `manifest.json` + `sha256`.

Manifest schema: [../docs/graph-pack-manifest.schema.json](../docs/graph-pack-manifest.schema.json).

## JSONL shapes (`ti`)

- `{"ioc":{...}}` — types: `ip`, `domain`, `url`, `hash`
- `{"campaign":{...}}` — optional `actors`, embedded `iocs`
- `{"cluster":{...}}` — optional nested `campaigns`
- `{"actor":{"name":"APT-X","description":"..."}}`
- `{"report":{"title":"...","provider":"...","link":"https://...","body_markdown":"..."}}`

## Stage 2 (future)

Kafka workers, STIX/MISP ingestion, Cue validation in CI, SBOM/Harbor, etc.
