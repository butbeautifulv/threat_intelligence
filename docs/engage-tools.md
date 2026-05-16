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
