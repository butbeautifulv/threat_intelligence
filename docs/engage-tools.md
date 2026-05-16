# Engage tool catalog

Tools are defined in YAML and loaded at startup (merged: `tools.yaml` ← `tools.live.yaml` ← `tools.enabled.yaml`).

| File | Purpose |
|------|---------|
| [tools.yaml](../engage/serve/catalog/tools.yaml) | Full catalog (~150 legacy MCP names); regenerate with `make catalog-engage` |
| [tools.live.yaml](../engage/serve/catalog/tools.live.yaml) | Default enabled tools for smoke (nmap, nuclei, httpx, subfinder, trivy) |
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

## Regenerate from reference

```bash
make catalog-engage          # scripts/engage/extract-legacy-catalog.py
make test-engage-parity      # 150 names vs .external/hexstrike_mcp.py
```

## Args templates

CLI `args` are generated when you run `make catalog-engage`:

| Source | Role |
|--------|------|
| `ARGS_TEMPLATES` in [extract-legacy-catalog.py](../scripts/engage/extract-legacy-catalog.py) | Explicit templates for **~100+** priority tools (network, web, osint, cloud, binary) |
| `INFER_TOOLS` + `infer_args_template()` | Heuristic templates from MCP parameter names (allowlist only) |
| Default | `["{target}", "{additional_args}"]` for all other catalog entries |

The API merges `target` into `url` / `domain` parameters when those fields exist on the tool spec. Runner expands placeholders via `BuildArgs` in [executor.go](../engage/serve/internal/runner/executor.go).

Priority tools with structured `args` are listed in `ARGS_TEMPLATES` (128+ keys as of Phase 9); after regen, the script prints how many tools have non-generic templates.

## CI tool matrix

`scripts/test/smoke-engage-tool-matrix.sh` exercises up to 10 tools when binaries exist on PATH: nmap, nuclei, httpx, subfinder, gobuster, nikto, ffuf, rustscan, trivy, sqlmap (best-effort skip).

## Enable by category (dev)

```bash
./scripts/engage/enable-catalog-by-category.sh network web
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

GitHub Actions workflow [`.github/workflows/engage.yml`](../.github/workflows/engage.yml) runs on push/PR when paths under `engage/`, `pkg/auth/`, `pkg/engage/`, or `scripts/engage/` change:

| Step | Command |
|------|---------|
| Unit tests + build | `make test-engage` |
| Catalog parity | `make test-engage-parity` |

`test-engage-parity` compares `tools.yaml` to `.external/hexstrike-ai-master/hexstrike_mcp.py`. Because `.external/` is gitignored, CI **skips** the strict name diff when that tree is absent (`skip parity` in the script). For a full 150-tool check locally, keep `.external/` on disk.

Tool execution smoke (`make test-engage-smoke-tool`) is not part of default CI; use `ENGAGE_SKIP_TOOL_SMOKE=1` or run it manually against a local `engage-api`.
