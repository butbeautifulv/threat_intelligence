# Scrapers (ingest → Neo4j)

Go services that pull public data and write into the shared Neo4j database. Built via [docker/](../docker/) Dockerfiles with repo root as context.

## Layout

| Directory | Binary | Role |
|-----------|--------|------|
| [vuln/](vuln/) | `vuln` | NVD CVE 2.0, Metasploit paths, Exploit-DB CSV, optional Vulners |
| [lola/](lola/) | `lola` | LOLBAS, GTFOBins, LOFTS, MITRE ATT&CK STIX enterprise |
| [ds/](ds/) | `ds` | SigmaHQ, YARA (Neo23x0), Atomic Red Team, Caldera Stockpile abilities |
| [ti/](ti/) | `ti` | CISA KEV, URLhaus, PT RSS, optional JSONL file |
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
| TI JSONL (local / mounted file) | `ti` | Implemented | — | `--input path.jsonl` (compose mounts [example.jsonl](ti/example.jsonl) as `/app/example.jsonl`) |

**TI feeds vs Docker:** root [docker-compose.yml](../docker-compose.yml) defaults `TI_FEEDS=kev,pt,urlhaus`. If the default PT URL returns HTML errors, use e.g. `TI_FEEDS=kev,urlhaus` or set `PT_RSS_URL` to a stable RSS endpoint you operate.

### Optional: same stack via Docker Compose

From repo root, `docker compose up --build` runs `neo4j` + `vuln` + `lola` + `ds` + `ti` + `panel` (see [../docker-compose.yml](../docker-compose.yml)). Re-run ingest without rebuild: `docker compose restart vuln lola ds ti`.

### Planned only (no parser here)

- AlienVault OTX API — future.
- MISP feeds — future.
- Abuse.ch ThreatFox / Feodo Tracker / MalwareBazaar (public CSV/API) — future; reference: `https://threatfox.abuse.ch/`, `https://feodotracker.abuse.ch/`, `https://bazaar.abuse.ch/`.

## Run locally (Neo4j must be up)

From repo root (paths under `scrapers/`):

```bash
cd scrapers/vuln && go run ./cmd
cd scrapers/lola && go run ./cmd
cd scrapers/ds && go run ./cmd
cd scrapers/ti && go run ./cmd --feeds kev,pt,urlhaus --input example.jsonl
```

`go.work` at repo root includes these modules; use `go work sync` if imports drift.

## Graph export and packs

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
