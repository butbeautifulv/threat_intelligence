# Threat Intelligence

## Sources (sites) intended for parsing

These are the original planned sources. The services in this repo ingest from the **public endpoints** listed below.

- **Vulns**:
  - NVD CVE API 2.0: `https://services.nvd.nist.gov/rest/json/cves/2.0`
  - Metasploit modules (GitHub): `https://github.com/rapid7/metasploit-framework/tree/master/modules`
  - Exploit-DB (public site): `https://www.exploit-db.com/`
  - Vulners (API/docs): `https://vulners.com/`
- **LOLA (Lolbins/LOLScripts)**:
  - LOLBAS (GitHub): `https://github.com/LOLBAS-Project/LOLBAS`
  - GTFOBins (GitHub): `https://github.com/GTFOBins/GTFOBins.github.io`
  - LOFTS (site): `https://lofts.galeal.com/`
  - MITRE ATT&CK (techniques): `https://attack.mitre.org/techniques/`
- **DS (Detection & Simulation)**:
  - Sigma rules (SigmaHQ GitHub): `https://github.com/SigmaHQ/sigma`
  - Atomic Red Team (GitHub): `https://github.com/redcanaryco/atomic-red-team`
  - Caldera (GitHub): `https://github.com/mitre/caldera`
- **TI (Artifacts / Reports)**:
  - CISA KEV feed (JSON): `https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json`
  - URLhaus recent (CSV): `https://urlhaus.abuse.ch/downloads/csv_recent/`
  - Positive Technologies RSS (configurable via `PT_RSS_URL`, default): `https://www.ptsecurity.com/rss/all.xml`

Planned coverage (sources of truth in **Neo4j** only):

- **vuln** — NVD CVE 2.0 (paginated, optional `NVD_API_KEY`, optional `NVD_MAX_PAGES` for smoke runs). Writes `Vulnerability`, `CWE`, `CPE` and `markdown` for the graph panel.
- **lola** — LOLBAS (`LOLBAS-Project/LOLBAS` `_lolbas/*.yml`) and GTFOBins (`GTFOBins/GTFOBins.github.io` `_gtfobins/*.md`). Writes `LolaArtifact`, `Command`, `markdown`. Raw files are cached under `LOLA_CACHE_DIR` (default `./data/cache`) so re-runs skip re-download when present.
- **ds** — SigmaHQ (sample of Windows process_creation rules), YARA (`Neo23x0/signature-base`, `yara` or `iocs/yara`), Atomic Red Team YAML. Writes `SigmaRule`, `YaraRule`, `AtomicTest` with `markdown`. Cache: `DS_CACHE_DIR` (default `./data/cache`).
- **ti** — JSONL envelopes (`ioc`, `campaign`, `cluster`, `actor`, `report`) plus optional public feeds: `kev` (CISA KEV), `pt` (PT RSS via `PT_RSS_URL`), `urlhaus` (recent CSV). Actors use stable `id` (hashed from name). Reports link to IOCs with `MENTIONS`.

Cue schemas: `cue_schemas/merge.cue` imports `schema/ds.cue` as the DS / `detect` bundle.

## One-command stack (Neo4j + ingest jobs + web panel)

```bash
docker compose up --build
```

- Neo4j Browser: `http://localhost:7474` (user `neo4j`, password `neo4jpassword`).
- **Graph panel (Obsidian-style markdown):** `http://localhost:8088` — force-directed graph; click a node to render `markdown` (or YAML of properties).

Environment knobs (examples):

- `NVD_MAX_PAGES` — limit NVD pagination for quick tests (compose default `1`).
- `TI_KEV_MAX`, `TI_PT_MAX`, `TI_URLHAUS_MAX` — cap feed volume.
- `GITHUB_TOKEN` — raises GitHub API rate limits for `lola` / `ds`.

## Run services locally (without Docker for Go)

Ensure Neo4j is up, then from repo root:

```bash
cd vuln && go run ./cmd
cd lola && go run ./cmd
cd ds && go run ./cmd
cd ti && go run ./cmd --feeds kev,pt,urlhaus --input example.jsonl
cd panel && npm i && npm run dev
```

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n) AS labels, count(*) AS c ORDER BY c DESC;
MATCH ()-[r]->() RETURN type(r) AS rel, count(*) AS c ORDER BY c DESC;
```

## JSONL shapes (`ti`)

- `{"ioc":{...}}` — types: `ip`, `domain`, `url`, `hash`
- `{"campaign":{...}}` — optional `actors`, embedded `iocs`
- `{"cluster":{...}}` — optional nested `campaigns`
- `{"actor":{"name":"APT-X","description":"..."}}`
- `{"report":{"title":"...","provider":"...","link":"https://...","body_markdown":"..."}}`

## Stage 2 (future)

Kafka workers, STIX/MISP ingestion, stronger schema validation (Cue in CI), SBOM/Harbor, etc.