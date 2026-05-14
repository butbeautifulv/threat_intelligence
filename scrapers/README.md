# Scrapers (ingest → Neo4j)

Go binaries that pull public data and write into the shared Neo4j database. Images are built from [docker/](../docker/) with the **repository root** as build context.

**Before a PR:** read [../docs/coding-style.md](../docs/coding-style.md) (layers, `slog`, NATS ingest).

---

## Layout

| Directory | Compose service / binary | Role |
|-----------|--------------------------|------|
| [vuln/](vuln/) | `vuln` | NVD CVE 2.0, Metasploit paths, Exploit-DB CSV, optional Vulners |
| [lola/](lola/) | `lola` | LOLBAS, GTFOBins, LOFTS, MITRE ATT&CK STIX enterprise |
| [ds/](ds/) | `ds` | SigmaHQ, YARA (Neo23x0), Atomic Red Team, Caldera Stockpile abilities |
| [ti/](ti/) | `ti` | CISA KEV, URLhaus, PT RSS, ThreatFox, MalwareBazaar, Feodo, OpenPhish, optional JSONL file |
| [sbom/](sbom/) | `sbom` | OSV per-CVE package ranges, GHSA JSON → `Package`, `SecurityAdvisory`, links to `Vulnerability` |
| [coderules/](coderules/) | `coderules` | MITRE CWE catalog (zip), Semgrep rules, sample CodeQL → `CWE`, `SemgrepRule`, `CodeQLRule` |
| [nuclei/](nuclei/) | `nuclei` (image binary `nuclei-scrape`) | [nuclei-templates](https://github.com/projectdiscovery/nuclei-templates) CVE YAML subset → `NucleiTemplate` |
| [ingest-worker/](ingest-worker/README.md) | **`ingest-worker`** | **Consumer:** JetStream pull on `ingest.>` → Neo4j (same writers as direct mode for AppSec, TI, vuln, lola, ds). See dedicated [ingest-worker README](ingest-worker/README.md). |
| [ingestpub/](ingestpub/) | *(library)* | Shared JetStream publisher + stream bootstrap (`INGEST` / `ingest.>`). |
| [proxybroker/](proxybroker/) | `proxybroker` | HTTP proxy pool for scrapers |
| [cue_schemas/](cue_schemas/) | — | Cue schemas (`merge.cue` imports `schema/ds.cue`) |

---

## ingest-worker (queue → Neo4j)

**`ingest-worker`** is a first-class scrape-profile service (see [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md#ingest-worker)). It does **not** fetch from the public internet; it **consumes** envelopes produced by scrapers when **`INGEST_MODE=nats`** (AppSec, TI, vuln, lola, ds).

| Topic | Doc |
|--------|-----|
| Env vars, local `go run`, Compose | [ingest-worker/README.md](ingest-worker/README.md) |
| Stream + dedup | [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md) (Compose service `nats`, `ingest-worker`) |
| Envelope schema | [../pkg/ingestv1/](../pkg/ingestv1/) |

---

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
| OSV API (per CVE) | `sbom` | Implemented | `https://api.osv.dev/v1/vulns/{CVE}` | `-sources osv`, `-max-cves` / `SBOM_MAX_CVES`; links `Vulnerability`→`Package` when CVE exists in graph |
| GitHub Advisory Database (raw JSON) | `sbom` | Implemented | `https://raw.githubusercontent.com/github/advisory-database/main/...` | `-sources ghsa`, `GITHUB_TOKEN` (recommended), `SBOM_MAX_GHSA`, `SBOM_GHSA_MIN_YEAR` |
| MITRE CWE catalog (zip) | `coderules` | Implemented | `https://cwe.mitre.org/data/xml/cwec_latest.xml.zip` | `CODERULES_MAX_CWE`; enriches `CWE` (`name`, `description`, `status` from MITRE) |
| Semgrep rules (GitHub API) | `coderules` | Implemented | `https://api.github.com/repos/semgrep/semgrep-rules/contents/...` | `CODERULES_MAX_SEMGREP`; optional `MAPS_TO_CWE` when rule metadata lists CWEs |
| CodeQL (GitHub API) | `coderules` | Implemented | `https://api.github.com/repos/github/codeql/contents/javascript/ql/src/Security/CWE-079` | `CODERULES_MAX_CODEQL`; `MAPS_TO_CWE` when `CWE-*` appears in query header |
| Nuclei templates (GitHub API) | `nuclei` | Implemented | `https://api.github.com/repos/projectdiscovery/nuclei-templates/contents/http/cves/...` | `NUCLEI_MAX`, `NUCLEI_YEARS`, `GITHUB_TOKEN` (recommended); `RELATES_TO_CVE`, `MAPS_TO_CWE` when metadata present |

**Ontology & roadmap:** [../docs/ontology-appsec.md](../docs/ontology-appsec.md).

---

## Optional NATS queue (`INGEST_MODE`)

Primary doc for the consumer: **[ingest-worker/README.md](ingest-worker/README.md)** and [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md#ingest-worker).

For **`sbom`**, **`coderules`**, **`nuclei`**, **`ti`**, **`vuln`**, **`lola`**, and **`ds`**, set **`INGEST_MODE=nats`** to publish versioned JSON envelopes to NATS JetStream instead of writing Neo4j from the scraper process. The scrape profile starts **`nats`** and **`ingest-worker`**; run the worker whenever producers use `nats` mode.

To drop **`NEO4J_*`** from producer containers in Compose while keeping **`ingest-worker`** on Bolt, add **`docker-compose.scrape-nats.yml`** (see [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md#nats-only-producers-optional-override)):

```bash
INGEST_MODE=nats docker compose -f docker-compose.yml -f docker-compose.scrape-nats.yml --profile scrape up --build -d
```

| Variable | Default | Meaning |
|----------|---------|--------|
| `INGEST_MODE` | `direct` | `direct` = write Neo4j from the scraper; `nats` = publish only (`coderules` / `nuclei` / **`sbom`** / **`ti`** / **`vuln`** / **`lola`** / **`ds`** skip Neo4j client when `nats`) |
| `NATS_URL` | `nats://localhost:4222` | Client URL; in Compose use `nats://nats:4222` |
| `SBOM_NATS_SUBJECT` | `ingest.appsec.sbom` | Publish subject for `sbom` |
| `SBOM_CVE_LIST_FILE` | *(Compose: `/fixtures/cve_list_seed.txt` in image)* | One `CVE-…` id per line (`#` comments allowed). Required for **`sbom`** OSV when `INGEST_MODE=nats` (replaces Neo4j `ListCVEs`). |
| `SBOM_CVE_LIST_URL` | empty | Alternative: HTTP(S) document with the same line format (used if **`SBOM_CVE_LIST_FILE`** is unset). |
| `CODERULES_NATS_SUBJECT` | `ingest.appsec.coderules` | Publish subject for `coderules` |
| `NUCLEI_NATS_SUBJECT` | `ingest.appsec.nuclei` | Publish subject for `nuclei` |
| `TI_NATS_SUBJECT` | `ingest.ti.events` | Publish subject for **`ti`** (`--feeds` / `--input` in `nats` mode) |
| `VULN_NATS_SUBJECT` | `ingest.vuln.events` | Publish subject for **`vuln`** |
| `LOLA_NATS_SUBJECT` | `ingest.lola.events` | Publish subject for **`lola`** |
| `DS_NATS_SUBJECT` | `ingest.ds.events` | Publish subject for **`ds`** |

Stream **`INGEST`**, subjects **`ingest.>`** (JetStream pattern; includes **`ingest.appsec.*`**, **`ingest.ti.*`**, …), dedup `Nats-Msg-Id` = envelope `idempotency_key`. Runtime env: [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md). **Envelope kinds, ack policy, and handler matrix:** [../docs/ingest-contract.md](../docs/ingest-contract.md).

**TI feeds vs Docker:** root [docker-compose.yml](../docker-compose.yml) defaults `TI_FEEDS=kev,urlhaus,threatfox,malwarebazaar,feodo`. Append `openphish` when you want phishing URLs (the remote feed can be slow or flaky; the scraper logs a warning and continues if the download fails). MalwareBazaar is skipped unless `MALWAREBAZAAR_AUTH_KEY` (or `MALWARE_BAZAAR_API_KEY`) is set. With `THREATFOX_AUTH_KEY`, ThreatFox uses the authenticated API instead of the public JSON export. Use `TI_PROXY_URLS` (and optional `TI_PROXY_MODE=only`) to route traffic via [proxybroker](proxybroker/). If the default PT URL returns HTML errors, add `pt` explicitly to `TI_FEEDS` only when needed, or set `PT_RSS_URL` to a stable RSS endpoint you operate.

---

## Docker Compose

From repo root, `docker compose up --build` runs **Neo4j** + **graph-bootstrap** + **HTTP API** (no live scraping). See [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md).

Re-run scrape profile services:

```bash
docker compose --profile scrape up --build -d
```

Restart ingest only (example):

```bash
docker compose restart vuln lola ds ti sbom coderules nuclei ingest-worker
```

**Graph maintenance:** [../scripts/README.md](../scripts/README.md) — `graph-dedup-cleanup.sh` (duplicate `HAS_ADVISORY`, optional stale isolated IOC removal; always `--dry-run` first).

---

## Planned only (no parser here)

- AlienVault OTX API — future.
- MISP feeds — future.

---

## Run locally (Neo4j must be up)

From repo root (paths under `scrapers/`):

```bash
cd scrapers/proxybroker && go run ./cmd
cd scrapers/vuln && go run ./cmd
cd scrapers/lola && go run ./cmd
cd scrapers/ds && go run ./cmd
cd scrapers/ti && go run ./cmd --feeds kev,urlhaus,threatfox,feodo,openphish --input example.jsonl
cd scrapers/sbom && go run ./cmd
cd scrapers/coderules && go run ./cmd
cd scrapers/nuclei && go run ./cmd
cd scrapers/ingest-worker && go run ./cmd
```

`proxybroker` does not talk to Neo4j; start it when scrapers need `*_PROXY_URLS`.

`go.work` at repo root includes `pkg/ingestv1`, `scrapers/ingestpub`, `scrapers/ingest-worker`, and scraper modules; run `go work sync` if imports drift.

---

## Graph export and packs

**Consuming a pack** (no scrapers): default Compose stack — `graph-bootstrap` imports a ZIP before `api` starts. Env vars: [../docs/threatintel-runtime.md](../docs/threatintel-runtime.md).

**Building a pack** (scraping): enable the **`scrape`** profile, fill Neo4j, then on the host:

- [../scripts/export-graph-cypher.sh](../scripts/export-graph-cypher.sh)
- [../scripts/build-graph-pack.sh](../scripts/build-graph-pack.sh) / [../scripts/import-graph-pack.sh](../scripts/import-graph-pack.sh)

Manifest schema: [../docs/graph-pack-manifest.schema.json](../docs/graph-pack-manifest.schema.json).

---

## JSONL shapes (`ti`)

- `{"ioc":{...}}` — types: `ip`, `domain`, `url`, `hash`; optional `sources` (string array) for provenance; `source` (single string) is still merged into `sources` and stored on the IOC node together with `firstSeen` / `lastSeen` timestamps (Neo4j requires APOC for merging `sources`; default Compose image enables APOC)
- `{"campaign":{...}}` — optional `actors`, embedded `iocs`
- `{"cluster":{...}}` — optional nested `campaigns`
- `{"actor":{"name":"APT-X","description":"..."}}`
- `{"report":{"title":"...","provider":"...","link":"https://...","body_markdown":"..."}}`

---

## IOC freshness and TTL (`ti`)

IOC nodes store **`firstSeen`**, **`lastSeen`**, **`sources`**, **`updatedAt`**, and legacy **`source`**. There is **no automatic expiry** in the write path: high-churn feeds (URLhaus, OpenPhish, ThreatFox) will accumulate nodes until you run an explicit cleanup policy.

**Recommended practice**

1. Use **`lastSeen`** (updated on every matching upsert) as the freshness clock.
2. Periodically run Cypher or [../scripts/graph-dedup-cleanup.sh](../scripts/graph-dedup-cleanup.sh) with `GRAPH_DELETE_STALE_ISOLATED_IOCS=1` **only after** reviewing counts — it removes **degree-0** IOCs older than `GRAPH_IOC_STALE_DAYS` (default 90) by comparing `coalesce(lastSeen, updatedAt)` to a UTC cutoff.
3. For campaign-linked IOCs, prefer targeted queries rather than blind global deletes.

---

## Stage 2 (future)

Kafka workers, STIX/MISP ingestion, Cue validation in CI, Harbor/binary SBOM, etc.
