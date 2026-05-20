# HexStrike (legacy) vs Veil API vs Engage — soft lab compare

For **HexStrike intelligence API probes directed at Veil URLs** (analyze-target / select-tools), see [hexstrike-pentest-veil.md](hexstrike-pentest-veil.md).

Lab-only run aligned with [external-hexstrike.md](external-hexstrike.md): legacy HexStrike is **reference**, not Veil production runtime. This note captures a **single gentle HTTP probe** per surface (no burst, no aggressive pentest).

## Topology (this run)

| Surface | Process / artifact | Listen | Purpose |
|---------|---------------------|--------|---------|
| HexStrike | `python3 hexstrike_server.py` in `.external/hexstrike-ai-master/` (venv `.venv`) | `8888` (env `HEXSTRIKE_PORT`; host env `HEXSTRIKE_HOST=127.0.0.1`) | Legacy Flask tool API |
| Veil API | `knowledge/serve/cmd/api` (`go run`) | `8090` | Graph read / `veil-api` REST |
| Engage | `engage/serve/cmd/api` via `run-client-native-api-instance.sh` | `8891` (victim) | Catalog + `POST /api/tools/{name}` execution layer |

Backing store for Veil API: Neo4j Bolt on `127.0.0.1:7687` (container `veil-local-neo4j` in this environment).

## Observed HTTP facts (soft probes)

| Endpoint | HTTP | Notes |
|----------|------|-------|
| `GET http://127.0.0.1:8888/health` | **200** | Large JSON; includes `all_essential_tools_available` and per-category availability stats. |
| `GET http://127.0.0.1:8090/health` | **200** | Body: `{"ok":true,"service":"veil-api"}`. |
| `GET http://127.0.0.1:8090/v1/categories` | **200** | Read-only; returns TI/vuln-style category list (first bytes only logged during run). |
| `GET http://127.0.0.1:8891/health` | **200** | Body includes `service":"veil-engage"` and `tool_count`: **158** (catalog size on server). |
| `GET http://127.0.0.1:8891/api/tools` | **200** | JSON `tools` array length in this response: **104** (shape may be filtered/paginated vs full catalog; health still reports 158). |

## Binding note (lab hygiene)

Werkzeug logged **“Running on all addresses (0.0.0.0)”** while also listing `127.0.0.1:8888`. Even with `HEXSTRIKE_HOST=127.0.0.1`, the dev server may still be reachable on other local interfaces. For any non-loopback network, **do not leave HexStrike running**; use a firewall or stop the process after the lab ([external-hexstrike.md](external-hexstrike.md)).

## Conceptual comparison

| Dimension | HexStrike (legacy) | Veil API (`veil-api`) | Engage (`veil-engage` API) |
|-----------|-------------------|----------------------|----------------------------|
| **Role** | Monolithic Flask “tool server” + workflows | Graph-backed read API (categories, graph queries, etc.) | Tool catalog + subprocess execution on host `PATH` |
| **Tool invocation** | Many Flask routes (per README / server) | Not the primary tool-runner for Veil agents | `POST /api/tools/{name}` + YAML catalog ([engage/README.md](../engage/README.md)) |
| **Auth (default legacy)** | No auth on MCP → HTTP path in reference topology ([external-hexstrike.md](external-hexstrike.md)) | Configurable (`AUTH_ENABLED`); this lab used open read checks | Same family of controls as hardened deploys; prod uses JWT/RBAC per [mcp-agents.md](mcp-agents.md) |
| **MCP story** | `hexstrike_mcp.py` stdio → HTTP `:8888` | `veil-mcp` (graph) separate | `veil-engage` MCP / `engage-api` — **agents should use Engage, not legacy HexStrike** ([mcp-agents.md](mcp-agents.md)) |

## How this differs from your Engage self-pentest layer

Your earlier **Engage-focused** self-pentest exercised **abuse resistance** and **prod-profile** expectations on Engage + graph MCP paths (see [engage-self-pentest-report.md](engage-self-pentest-report.md)). This document is intentionally **minimal**: it only proves **coexistence** and **basic HTTP health/read** between legacy HexStrike and the Veil split (`veil-api` + `veil-engage`) for side-by-side understanding — not a second aggressive pass.

## Remediation / ops hints

- Prefer **no** long-lived HexStrike on shared laptops without localhost firewall rules.
- For product comparisons, rely on **parity and benchmark scripts** (`scripts/benchmark/engage-hexstrike-parity.sh`, `make test-engage-parity`) rather than running Flask in prod-like environments.

## Teardown (this session)

- **Stopped:** listener on `127.0.0.1:8888` (HexStrike Flask dev server) after probes.
- **Left running:** `veil-api` on `:8090`, Engage on `:8891`, and Docker `veil-local-neo4j` were already present before this lab; they were **not** stopped to avoid breaking your active dev stack. Stop them manually when finished (`pkill` / `docker stop veil-local-neo4j` as appropriate).
