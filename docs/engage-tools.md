# Engage tool catalog

**Navigation:** [engage-runtime.md](engage-runtime.md) · [engage-install-linux.md](engage-install-linux.md) · [engage-lab-pentest.md](engage-lab-pentest.md) · [engage-legacy-parity.md](engage-legacy-parity.md) · generated [engage-tools-na-matrix.md](engage-tools-na-matrix.md) · [engage-executable-gaps.md](engage-executable-gaps.md) · [engage-tool-install-coverage.md](engage-tool-install-coverage.md)

Tools are defined in YAML and loaded at startup (merged in order, **later overrides**): `tools.yaml` → `tools.live.yaml` → `tools.enabled.yaml`.

## Three KPIs (do not conflate)

| KPI | Today | Target | Gate |
|-----|-------|--------|------|
| **Catalog** | **158** names in `tools.yaml` | 158 | `make test-engage-parity` |
| **Executable** (callable via HTTP+MCP) | Partial (~55 bridge + subset subprocess) | **158/158** | `make test-engage-executable-matrix` (**P9f**) |
| **Subprocess-in-runner** | **~46** catalog names `enabled: true`; **136** live rows total | **~103** subprocess enabled, 0 orphan rows | **P9h** (catalog ↔ live sync) + **P9i** (binaries in runner) |

**Phase 30 sign-off** ([engage-audit-report.md](engage-audit-report.md)) = decommission legacy `:8888` + catalog/route parity — **not** “every catalog tool runs in the default runner.”

## Orphan live rows (P9h)

`tools.live.yaml` currently has **136** `enabled: true` entries but only **~46** names exist in `tools.yaml` (**catalog ∩ live**). The other **~90** are **orphan synthetics** (e.g. `nmap_quick_scan` clones) left from older `generate-tools-live.py` policy. They inflate “live” counts in docs and matrix headers without matching MCP catalog names.

**P9h fix:** one row per catalog name (158), drop non-catalog synthetics, set subprocess `enabled: true` only when the binary is in the runner profile. Regenerate: `make catalog-engage` after overlay script lands. See [engage_tools_full_coverage.plan.md](../.cursor/plans/engage_tools_full_coverage.plan.md).

| Count | Meaning |
|-------|---------|
| **158** | Every tool **name** in `tools.yaml` (MCP parity + bridge aliases) |
| **~55** | **bridge_api** — in-process intel / CTF / bug bounty / workflows (not subprocess) |
| **~57** | **runner_N/A** — CLI in catalog but missing from runner image or not yet enabled (**P9i**) |
| **~12** | P9g heavy stack (Burp, Ghidra, hashcat, …) — subprocess via `engage-runner-full` |

| File | Purpose |
|------|---------|
| [tools.yaml](../engage/serve/catalog/tools.yaml) | Full catalog (158 names); regenerate with `make catalog-engage` |
| [tools.live.yaml](../engage/serve/catalog/tools.live.yaml) | Lab profile: subprocess tools `enabled: true`; regen via `generate-tools-live.py` |
| [engage-tools-na-matrix.md](engage-tools-na-matrix.md) | Execution status per name (`live` / `runner_N/A` / `bridge_api`) |
| [engage-llm-stubs.md](engage-llm-stubs.md) | `ai_*` catalog tools: deterministic stub vs future LLM; `tryAgentTool` / intel bridge dispatch |
| [tools.enabled.yaml](../engage/serve/catalog/tools.enabled.yaml) | Auto-generated enables when binaries exist on PATH |

## Schema

Each entry includes:

| Field | Description |
|-------|-------------|
| `name` | Stable tool id (e.g. `nmap_scan`) — matches legacy MCP |
| `category` | `network`, `web`, `cloud`, `binary`, `auth`, `osint`, `ctf`, `intelligence` |
| `binary` | Executable on PATH (or in engage-runner image) |
| `parameters` | Input fields (name, type, default, required) — drives MCP `inputSchema` |
| `args` | CLI template with `{target}`, `{scan_type}`, `{ports}`, etc. |
| `timeout_sec` | Subprocess timeout |
| `enabled` | `true` only when binary is available |

## Runner tool matrix (Phase 19)

Image: [runner.Dockerfile](../deploy/engage/docker/runner.Dockerfile). List binaries: `./scripts/engage/list-runner-binaries.sh`.

| Binary | In runner | Example catalog id | Live enabled |
|--------|-----------|-------------------|--------------|
| nmap | yes | `nmap_scan` | yes |
| masscan | yes | `masscan_high_speed` | yes |
| nuclei | yes | `nuclei_scan` | yes |
| httpx | yes | `httpx_probe` | yes |
| subfinder | yes | `subfinder_scan` | yes |
| amass | yes | `amass_scan` | yes |
| katana | yes | `katana_crawl` | yes |
| gau | yes | `gau_discovery` | yes |
| waybackurls | yes | `waybackurls_discovery` | yes |
| gobuster | yes | `gobuster_scan` | yes |
| feroxbuster | yes | `feroxbuster_scan` | yes |
| ffuf | yes | `ffuf_scan` | yes |
| nikto | yes | `nikto_scan` | yes |
| sqlmap | yes | `sqlmap_scan` | yes |
| dalfox | yes | `dalfox_xss_scan` | yes |
| arjun | yes | `arjun_scan` | yes |
| dirsearch | yes | `dirsearch_scan` | yes |
| paramspider | yes | `paramspider_discovery` | yes |
| rustscan | yes | `rustscan_fast_scan` | yes |
| trivy | yes | `trivy_scan` | yes |
| naabu | yes | `naabu_port_scan` (live synthetic) | yes |
| dnsx | yes | `dnsx_resolve` (live synthetic) | yes |
| dnsenum, fierce, hydra, wafw00f, enum4linux, dirb | yes | `*_scan` | yes |

Compose runner profile uses `tools.live.yaml` ([compose.runner.yml](../deploy/engage/compose.runner.yml)).

## Regenerate from reference

```bash
make catalog-engage          # extract-legacy-catalog.py + generate-tools-live.py
make test-engage-parity      # 150 names vs .external/hexstrike_mcp.py
make test-engage-catalog-args  # 150/150 args templates or documented generic
```

## Args templates

CLI `args` are generated when you run `make catalog-engage`:

| Source | Role |
|--------|------|
| `ARGS_TEMPLATES` in [extract-legacy-catalog.py](../scripts/engage/extract-legacy-catalog.py) | Explicit templates for priority tools |
| `INFER_TOOLS` + `infer_args_template()` | Heuristic templates from MCP parameter names |
| `CATEGORY_ARGS` | Category-specific defaults (web `-u`, osint `-d`, …) |
| `DOCUMENTED_GENERIC` | In-process / workflow tools that intentionally use generic args |

Gate: [check-catalog-args.sh](../scripts/engage/check-catalog-args.sh) — fails CI if any tool has undocumented generic args.

## Heavy stack (full port — P9g)

Default `runner.Dockerfile` target `engage-runner` — tier-1 CLI only. Target **`engage-runner-full`** adds headless wrappers (no GUI):

| Binary | Catalog tools | RAM note |
|--------|---------------|----------|
| burpsuite | `burpsuite_scan`, `burpsuite_alternative_scan` | JRE headless |
| ghidra | `ghidra_analysis` | +2–4 GB during analyzeHeadless |
| hashcat, john | `hashcat_crack`, `john_crack` | GPU optional; CPU wordlists |
| gdb | `gdb_analyze`, `gdb_peda_debug` | batch mode |
| metasploit | `metasploit_run` | `msfconsole -q -x` |
| angr | `angr_symbolic_execution` | pip; memory-heavy |
| radare2 | `radare2_analyze` | wraps `r2` |
| volatility | `volatility_analyze` | volatility3 / `vol` |
| wpscan | `wpscan_analyze` | Ruby gem |

Image **~8–12 GB** on disk; allocate **4–8 GB RAM** for the full runner container. See [deploy/engage/README.md](../deploy/engage/README.md).

```bash
docker compose -f deploy/engage/compose.yml --profile runner-full up -d --build engage-runner-full
ENGAGE_RUNNER_PROFILE=full ENGAGE_RUNNER_IMAGE=engage-runner-full ./scripts/engage/list-runner-binaries.sh
```

`binary: api` / workflow placeholders remain **bridge**, not subprocess.

Full matrix: [engage-tools-na-matrix.md](engage-tools-na-matrix.md) (`make test-engage-na-matrix`).

## CI tool matrix

[tool-matrix-from-effectiveness.py](../scripts/engage/tool-matrix-from-effectiveness.py) builds [tool-matrix.targets](../scripts/engage/tool-matrix.targets) from tools with effectiveness ≥ 0.85.

```bash
make test-engage-tool-matrix   # best-effort; ≥15 on CI, ≥30 in runner profile (ENGAGE_TOOL_MATRIX_STRICT=1)
```

## Enable by category (dev)

```bash
./scripts/engage/enable-tools-on-path.sh    # default: network web osint cloud binary
# writes engage/serve/catalog/tools.enabled.yaml
```

## API request shape

```json
{
  "target": "example.com",
  "additional_args": "-T4",
  "parameters": {
    "scan_type": "-sV",
    "ports": "80,443"
  }
}
```

Runner merges `parameters` with catalog defaults, then expands `args` templates. See [engage-runtime.md](engage-runtime.md) for `ENGAGE_RUNNER_MODE=local|docker`.

## CI

GitHub Actions [engage.yml](../.github/workflows/engage.yml):

| Step | Command |
|------|---------|
| Unit tests + build | `make test-engage` |
| Catalog parity | `make test-engage-parity` |
| Args gate | `make test-engage-catalog-args` |
| Tool matrix | `smoke-engage-tool-matrix.sh` |

Bug bounty execute smoke: `scripts/test/smoke-bugbounty-recon-execute.sh` (included in `make test-engage-bugbounty`).

## Benchmark regression (not KPI)

Benchmark targets record **timing artifacts for regression** when engage-api (and usually docker runner) are up locally or in nightly jobs. They are **not** product KPIs, SLA gates, or marketing claims (no “24× faster” interpretation).

| Target | Script | Endpoints / tools |
|--------|--------|-------------------|
| `make test-engage-benchmark` | [engage-hexstrike-parity.sh](../scripts/benchmark/engage-hexstrike-parity.sh) | Workflow: recon, smart-scan, assessment-report |
| `make test-engage-benchmark-regression` | [engage-benchmark-baseline.sh](../scripts/test/engage-benchmark-baseline.sh) | Tool runs: `nmap_scan`, `nuclei_scan` |

Both write JSON under `scripts/benchmark/results/` (`latest.json`, `baseline-tools.json`) — gitignored. SKIP or stub JSON is OK when API, curl, or docker is unavailable. Neither target is required in PR CI; use them to compare runs over time, not to pass/fail releases on seconds alone.

## Assessment reports

Export branded deliverables via `POST /api/visual/export-report`.

```json
{
  "target": "https://example.com",
  "format": "html",
  "findings": [{"title": "Open port 443", "severity": "medium"}],
  "branding": {
    "organization": "Acme Security",
    "classification": "CONFIDENTIAL",
    "footer": "Authorized assessment only"
  }
}
```

| Field | Values |
|-------|--------|
| `format` | `pdf` (default) or `html` |
| `branding` | Optional org name, classification banner, footer |
| `summary_report` | Alternative to inline `target` + `findings` |

**Response:** PDF → `pdf_base64`, `size_bytes`; HTML → `html` body. Set `save_file: true` to persist under `ENGAGE_FILES_DIR`.

**Structured findings** (assessment / smart-scan): nuclei JSONL, nmap open ports, ffuf/sqlmap shapes — see parsers in [engage-llm-stubs.md](engage-llm-stubs.md) for `ai_*` vs subprocess tools.
