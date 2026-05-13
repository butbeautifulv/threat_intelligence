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

| Source | Service | Parser status | Notes |
|--------|---------|----------------|-------|
| NVD CVE API 2.0 | `vuln` | Implemented | `NVD_API_KEY`, `NVD_MAX_PAGES`, `VULN_REQUEST_DELAY`, `VULN_CACHE_DIR` |
| Metasploit modules (GitHub) | `vuln` | Implemented | `GITHUB_TOKEN`, `VULN_METASPLOIT_MAX_RB`, `VULN_METASPLOIT_MAX_DIRS` |
| Exploit-DB CSV (GitLab mirror) | `vuln` | Implemented | `VULN_EXPLOITDB_CSV_URL`, `VULN_EXPLOITDB_MAX_ROWS` |
| Vulners | `vuln` | Implemented when key set | `VULNERS_API_KEY`; `VULN_VULNERS_MAX`, `VULN_VULNERS_QUERY` |
| LOLBAS / GTFOBins | `lola` | Implemented | `LOLA_CACHE_DIR`, `GITHUB_TOKEN` |
| LOFTS | `lola` | Implemented | `LOFTS_URL`, `LOFTS_SKIP_ON_ERROR`, `LOFTS_MAX_ENTRIES` |
| MITRE ATT&CK (STIX) | `lola` | Implemented | `LOLA_MITRE_STIX_URL`, `LOLA_MITRE_MAX_TECHNIQUES`, `LOLA_MITRE_MAX_RELATIONSHIPS` |
| Sigma / YARA / ART / Caldera | `ds` | Implemented | `DS_MAX_*`, `DS_CACHE_DIR`, `GITHUB_TOKEN`, `DS_MAX_CALDERA`, `DS_CALDERA_BASE_PATH` |
| CISA KEV / URLhaus / PT RSS | `ti` | Implemented | `TI_*_MAX`, `TI_FEEDS`, `TI_CACHE_DIR` |
| TI JSONL | `ti` | Implemented | `--input path.jsonl` |

**TI feeds vs Docker:** root [docker-compose.yml](../docker-compose.yml) defaults `TI_FEEDS=kev,pt,urlhaus`. Override with `TI_FEEDS=kev,urlhaus` to drop PT RSS.

### Planned only (no parser here)

- AlienVault OTX API — future.
- MISP feeds — future.
- Abuse.ch Feodo / MalwareBazaar — future.

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
