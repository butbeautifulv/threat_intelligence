# Scrape layer

Fetches third-party feeds, records fetches in the Vitess/MySQL **ledger**, and publishes **`scrapev1`** envelopes to NATS **`scrape.>`**.

## Layout

| Path | Role |
|------|------|
| [scrape_worker/](scrape_worker/) | Entry binary; runs [factory](../scrape/factory/) sources |
| [sources/](sources/) | Per-domain scrapers (`ti`, `vuln`, `lola`, `ds`, `sbom`, `coderules`, `nuclei`) |
| [factory/](factory/) | Source registry, `RunAll`, continue-on-error unless `SCRAPE_FAIL_FAST=1` |
| [feeds/](feeds/) | Shared HTTP/GitHub fetch (codeload-first, tree API fallback) |
| [ledger/](ledger/) | `crawl_resource` dedup / refetch policy |
| [pub/](pub/) | JetStream publish helpers |
| [contract/scrapev1/](contract/scrapev1/) | Envelope kinds and DTOs |

**Build:** `cd scrape && go build ./scrape_worker/...`  
**Deploy:** [deploy/scrape/compose.yml](../deploy/scrape/compose.yml)

## Sources (`SCRAPE_SOURCES`)

Default (compose): `ds,vuln,lola,ti,sbom,coderules,nuclei`

| Source | Publishes | Notable env |
|--------|-----------|-------------|
| `vuln` | NVD pages, Exploit-DB, optional Vulners/MSF | `NVD_MAX_PAGES`, `VULN_EXPLOITDB_MAX_ROWS`, `VULN_METASPLOIT_MAX_RB`, `NVD_API_KEY` |
| `lola` | LOLBAS, GTFOBins, LOFTS, MITRE STIX | `LOLA_MITRE_MAX_TECHNIQUES`, `LOFTS_SKIP_ON_ERROR`, `LOFTS_URL` |
| `ds` | Sigma, YARA, Atomic Red Team, Caldera | `DS_MAX_SIGMA`, `DS_MAX_YARA`, `DS_MAX_ATOMIC`, `DS_MAX_CALDERA` |
| `ti` | KEV, ThreatFox, MalwareBazaar, Feodo, JSONL | `TI_FEEDS`, `TI_JSONL_FILE` |
| `sbom` | OSV, GHSA | `SBOM_SOURCES`, `SBOM_MAX_CVES`, `SBOM_MAX_GHSA`, `SBOM_CVE_LIST_FILE` |
| `coderules` | Semgrep, CodeQL, CWE catalog | `CODERULES_MAX_SEMGREP`, `CODERULES_MAX_CODEQL` |
| `nuclei` | Nuclei templates | `NUCLEI_MAX` |

Ledger: `VITESS_DSN`, `SCRAPE_MIN_REFETCH_AFTER`, `SCRAPE_FORCE_REFETCH=1` for full refetch.

## Local run

Requires NATS (and ledger DB for periodic sources):

```bash
export NATS_URL=nats://localhost:4222
export SCRAPE_SOURCES=vuln
export NVD_MAX_PAGES=1
export GRAPH_PACK_SKIP=1
cd scrape/scrape_worker && go run .
```

Full stack: `./scripts/compose-up-full.sh` from repo root.

## Graph packs

Scrape fills Neo4j via pipeline + ingest; export on the host:

```bash
./scripts/export-graph-cypher.sh
GRAPH_PACK_VERSION=v0.3.2 ./scripts/build-graph-pack.sh
```

Fast-rich profile: [scripts/graph-pack-run-v032.sh](../scripts/graph-pack-run-v032.sh).

## Further reading

- [docs/threatintel-runtime.md](../docs/threatintel-runtime.md) — ports, NATS subjects, compose
- [docs/ingest-contract.md](../docs/ingest-contract.md) — `scrapev1` kinds
- [pipeline/README.md](../pipeline/README.md) — downstream normalization
