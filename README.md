# Threat Intelligence

## Sources matrix (ingest → Neo4j)

| Source | Service | Parser status | Notes |
|--------|---------|----------------|-------|
| NVD CVE API 2.0 | `vuln` | Implemented | `NVD_API_KEY`, `NVD_MAX_PAGES`, `VULN_REQUEST_DELAY`, `VULN_CACHE_DIR` |
| Metasploit modules (GitHub) | `vuln` | Implemented | `GITHUB_TOKEN`, `VULN_METASPLOIT_MAX_RB`, `VULN_METASPLOIT_MAX_DIRS` |
| Exploit-DB CSV (GitLab mirror) | `vuln` | Implemented | `VULN_EXPLOITDB_CSV_URL`, `VULN_EXPLOITDB_MAX_ROWS` |
| Vulners | `vuln` | Implemented when key set | `VULNERS_API_KEY` required; `VULN_VULNERS_MAX`, `VULN_VULNERS_QUERY`; no-op + log if key empty |
| LOLBAS (GitHub) | `lola` | Implemented | `LOLA_CACHE_DIR`, `GITHUB_TOKEN` |
| GTFOBins (GitHub) | `lola` | Implemented | same |
| LOFTS | `lola` | Implemented | `LOFTS_URL` (default `https://lofts.galeal.com/`), `LOFTS_SKIP_ON_ERROR=true` to continue on HTTP errors |
| MITRE ATT&CK (STIX enterprise) | `lola` | Implemented | `LOLA_MITRE_STIX_URL`, `LOLA_MITRE_MAX_OBJECTS`; links `LolaArtifact` / `Command` where `mitreID` matches technique `id` |
| SigmaHQ (GitHub) | `ds` | Implemented | `DS_MAX_SIGMA`, `DS_CACHE_DIR`, `GITHUB_TOKEN` |
| YARA Neo23x0 (GitHub) | `ds` | Implemented | `DS_MAX_YARA`, `DS_YARA_PATH` |
| Atomic Red Team (GitHub) | `ds` | Implemented | `DS_MAX_ATOMIC` |
| Caldera abilities (GitHub) | `ds` | Implemented | `DS_MAX_CALDERA`, `DS_CALDERA_ABILITIES_PATH` |
| CISA KEV JSON | `ti` | Implemented | `TI_KEV_MAX`, `TI_CACHE_DIR` |
| URLhaus recent CSV | `ti` | Implemented | `TI_URLHAUS_MAX` |
| Positive Technologies RSS | `ti` | Implemented | `PT_RSS_URL`, `TI_PT_MAX`; include `pt` in `TI_FEEDS` (see below) |
| TI JSONL (`ioc`, `campaign`, …) | `ti` | Implemented | `--input path.jsonl` |

**TI feeds vs Docker:** [docker-compose.yml](docker-compose.yml) uses `TI_FEEDS` (default `kev,pt,urlhaus`) so compose matches the full public TI set. Override with `TI_FEEDS=kev,urlhaus` if you want to skip PT RSS.

### Planned / declared only (no parser in this repo)

These URLs are tracked for roadmap context; they are **not** ingested by the services above.

- AlienVault OTX pulses (API): `https://otx.alienvault.com/api/v1/` — requires API key; future worker.
- MISP feed block (instance-specific): document your `MISP_BASE/feeds/` URL in deployment config; future STIX/MISP path (see Stage 2).
- Abuse.ch Feodo Tracker CSV: `https://feodotracker.abuse.ch/downloads/ipblocklist.csv` — future `ti` feed.
- MalwareBazaar (abuse.ch): `https://bazaar.abuse.ch/` — future TI artifacts.

Cue schemas: `cue_schemas/merge.cue` imports `schema/ds.cue` as the DS / `detect` bundle.

## One-command stack (Neo4j + ingest jobs + web panel)

```bash
docker compose up --build
```

- Neo4j Browser: `http://localhost:7474` (user `neo4j`, password `neo4jpassword`). APOC is enabled for optional Cypher export (see below).
- **Graph panel (Obsidian-style markdown):** `http://localhost:8088` — force-directed graph; click a node to render `markdown` (or YAML of properties).

Environment knobs (examples):

- `NVD_MAX_PAGES` — limit NVD pagination for quick tests (compose default `1`).
- `TI_KEV_MAX`, `TI_PT_MAX`, `TI_URLHAUS_MAX` — cap feed volume.
- `GITHUB_TOKEN` — raises GitHub API rate limits for `lola` / `ds` / `vuln` (Metasploit tree).

## Export graph to a file (Neo4j / APOC)

After data is loaded, generate a portable Cypher script (for another Neo4j or cold import):

```bash
./scripts/export-graph-cypher.sh
```

Output: `./data/neo4j_export/graph.cypher` (inside the container this is under the Neo4j `import` tree: `import/user_export/graph.cypher`).

Load elsewhere:

```bash
cat data/neo4j_export/graph.cypher | docker run --rm -i \
  -e NEO4J_URI=neo4j://host.docker.internal:7687 \
  neo4j:5 cypher-shell -u neo4j -p '<password>' -d neo4j
```

(`CALL apoc.export.cypher.all` takes the relative path under the Neo4j `import` tree and an optional config map; the active database is the one selected in `cypher-shell` with `-d`.)

**Alternative (binary):** from the Neo4j container, `neo4j-admin database dump neo4j --to-path=/backups` after stopping the DB — good for same-major restore; see Neo4j ops docs.

### Graph release packs (offline / “ruleset” style)

Workflow similar to shipping a Semgrep rules bundle: you **build once** (slow, respectful rate limits), publish a **versioned ZIP** on GitHub Releases, and air‑gapped installs **only import** — no `vuln`/`lola`/`ds`/`ti` scraping.

1. **Saturate** Neo4j (compose or local), tuning delays and tokens as needed.
2. **Export** Cypher: `./scripts/export-graph-cypher.sh`
3. **Pack** (manifest + checksum + `graph.cypher` in one zip):

```bash
GRAPH_PACK_VERSION=v2026.05.0 ./scripts/build-graph-pack.sh
# or re-export then pack in one step:
EXPORT_FIRST=1 GRAPH_PACK_VERSION=v2026.05.0 ./scripts/build-graph-pack.sh
```

Artifact: `data/neo4j_export/releases/threat-intel-graph-<version>.zip` plus a copy `manifest.<version>.json`. Manifest schema id: `threat-intelligence.graph-pack/1` (see [docs/graph-pack-manifest.schema.json](docs/graph-pack-manifest.schema.json)).

4. **Publish** the zip on a GitHub Release (attach asset; do not commit large binaries to git by default — `data/neo4j_export/releases/` is gitignored).

5. **Import** on an autonomous host (Neo4j up; unpack verifies `sha256` then streams statements):

```bash
export NEO4J_PASS='...'
./scripts/import-graph-pack.sh ./data/neo4j_export/releases/threat-intel-graph-v2026.05.0.zip
# or from a release URL:
./scripts/import-graph-pack.sh 'https://github.com/<org>/<repo>/releases/download/<tag>/threat-intel-graph-v2026.05.0.zip'
```

With Docker Compose from repo root (runs `cypher-shell` inside the `neo4j` service):

```bash
export NEO4J_URI=neo4j://localhost:7687
export USE_DOCKER_COMPOSE=1
./scripts/import-graph-pack.sh /path/to/threat-intel-graph-v2026.05.0.zip
```

Requirements for scripts: `zip`, `unzip`, `python3`, `curl` (for HTTPS packs). Target DB should be **Neo4j 5.x** (same line as export).

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
