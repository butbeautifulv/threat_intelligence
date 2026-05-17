# Engage tool catalog

Tools are defined in YAML and loaded at startup (merged in order, **later overrides**): `tools.yaml` → `tools.live.yaml` → `tools.enabled.yaml`.

| File | Purpose |
|------|---------|
| [tools.yaml](../engage/serve/catalog/tools.yaml) | Full catalog (~150 legacy MCP names); regenerate with `make catalog-engage` |
| [tools.live.yaml](../engage/serve/catalog/tools.live.yaml) | Lab profile: **100+** tier-1 tools enabled; regen via `generate-tools-live.py` |
| [engage-tools-na-matrix.md](engage-tools-na-matrix.md) | Execution status for all 158 catalog names (`live` / `runner_N/A` / `bridge_api` / `permanent_N/A`) |
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

## Permanent N/A (Phase 25)

These stay **out of engage-runner** by design (GUI, multi-GB, or legacy-only stacks):

| Tool / family | Reason |
|---------------|--------|
| wpscan | Ruby gem stack; use `nikto` / `nuclei` in runner |
| ghidra, burpsuite, metasploit, angr | GUI or heavy frameworks |
| `binary: api` / workflow placeholders | In-process bridge, not subprocess |

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
