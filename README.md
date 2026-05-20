# Veil (Vulnerability Exploitation Intelligence Layer)

![Veil](docs/veil.png)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Veil** is a Neo4j-backed threat-intelligence platform with an optional **active security testing** layer. The graph holds CVE/CWE/CPE, LOLbins-style artifacts, detection content (Sigma/YARA/Caldera), TI feeds, SBOM advisories, and code-rule templates. Runtime is **four isolated Go modules** — **discovery**, **pipeline**, **knowledge** (read intel), **engage** (tool execution) — on **NATS JetStream** for ingestion and **dual MCP** servers for agents.

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Agents / AI:** [AGENTS.md](AGENTS.md) · **Security:** [SECURITY.md](SECURITY.md) · **Code of conduct:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

---

## Current state (2026-05-20)

| Area | Status | Details |
|------|--------|---------|
| **Platform** | P0–P12 + v8 **done** | Unified edge `veil-edge` — [platform-unified-access.md](docs/platform-unified-access.md); history in [.cursor/plans/archive/README.md](.cursor/plans/archive/README.md) |
| **Engage sign-off** | Phase 30 **done** | Legacy HexStrike `:8888` decommissioned (architecture + routes) — [engage-audit-report.md](docs/engage-audit-report.md). **Not** the same as 158/158 subprocess execution |
| **Active tracks** | In progress | **P9f** `make test-engage-executable-matrix`; plans: [engage_tools_full_coverage](.cursor/plans/engage_tools_full_coverage.plan.md), [client-native master](.cursor/plans/engage_mcp_client_native_execution_master.plan.md) |
| **Catalog** | **158** names | 151 legacy MCP + 8 engage bridge — [engage-tools.md](docs/engage-tools.md) |
| **Host install (lab)** | Core47 **46/47** | `ghidra` manual — [engage-install-linux.md](docs/engage-install-linux.md) |
| **Lab security** | Prod pentest **No-Go** | 2× high: open `/api/tools`, `/v1/categories` — [engage-lab-pentest.md](docs/engage-lab-pentest.md), [engage-red-blue-bugs.md](docs/engage-red-blue-bugs.md) |

---

## What you get

| Capability | Description |
|------------|-------------|
| **Threat graph** | Versioned [graph packs](docs/graph-pack.md), HTTP API (`/v1/*`), read-only MCP |
| **Ingestion bus** | Scrape → NED → ingest over NATS (`pkg/harvest`, `pkg/commit`) |
| **Engage toolkit** | **158** catalog tools · bridge + subprocess matrix · **client-native host PATH** by default |
| **Closed loop** | Tool runs → `engage.events` → graph (`EngageToolRun`, `EngageFinding`) |
| **Unified edge (P12)** | One TLS host: `/v1/*`, `/api/*`, `/mcp/graph`, `/mcp/engage`; scale **4 / 8 / 16** |
| **Agent-ready** | **veil-mcp** (read) + **veil-engage** (exec), Keycloak RBAC, GAIA eval harness |
| **Prod path** | Terraform + Ansible + Helm; [veil-controls](deploy/security/veil-controls.yaml) |

---

## Architecture

System diagram and layer roles: see mermaid in [docs/platform-architecture.md](docs/platform-architecture.md) (or the summary below).

| Layer | Path | Role | MCP |
|-------|------|------|-----|
| **Discovery** | [discovery/](discovery/) | Feeds, Vitess ledger, `harvest` publish | — |
| **Pipeline** | [pipeline/](pipeline/) | NED → `commit`; [engage-events](pipeline/engage-events/) | — |
| **Knowledge** | [knowledge/](knowledge/) | Neo4j ingest + [serve](knowledge/serve/) API/MCP | `veil-mcp` (read) |
| **Engage** | [engage/](engage/) | Catalog tools, workflows, reports | `veil-engage` (exec) |

**Shared `pkg/`:** [harvest](pkg/harvest/), [commit](pkg/commit/), [natsjet](pkg/natsjet/), [auth](pkg/auth/), [engage](pkg/engage/), [report](pkg/report/), [decision](pkg/decision/), [exec](pkg/exec/), [api](pkg/api/), [mcp](pkg/mcp/). **No Go imports** across the four layer roots.

**Agents:** **veil-mcp** + **veil-engage** only — [mcp-agents.md](docs/mcp-agents.md). Legacy HexStrike is reference-only — [external-hexstrike.md](docs/external-hexstrike.md).

**Contracts:** [ingest-contract.md](docs/ingest-contract.md) · [threatintel-runtime.md](docs/threatintel-runtime.md) (ports) · [engage-runtime.md](docs/engage-runtime.md) · [deploy/](deploy/)

---

## Quick start

Compose under [deploy/](deploy/); presets in [deploy/stacks/](deploy/stacks/).

### Graph only (Neo4j + API + optional MCP)

```bash
docker compose -f deploy/knowledge/compose.yml up -d --build
docker compose -f deploy/knowledge/compose.yml --profile mcp up -d --build mcp   # optional
curl -sS http://localhost:8090/health
make test-graph-read-smoke
```

Pack version: [versions.env](versions.env) → `GRAPH_PACK_VERSION` (currently **v0.4.6**).

### Unified edge (graph + engage, TLS)

[platform-unified-access.md](docs/platform-unified-access.md) · `make test-platform-unified-edge`

### Engage (host PATH default)

```bash
docker compose -f deploy/engage/compose.yml up -d --build engage-api engage-mcp
curl -sS http://localhost:8890/health | jq .
```

**Execution:** subprocesses on the **MCP host `PATH`** ([engage-mcp-topology.md](docs/engage-mcp-topology.md)). Optional **runner** profile = docker-exec lab only — [deploy/engage/README.md](deploy/engage/README.md).

Install CLIs: [engage-install-linux.md](docs/engage-install-linux.md) · Lab pentest notes: [engage-lab-pentest.md](docs/engage-lab-pentest.md).

### Full scrape pipeline

```bash
./scripts/ops/compose-up-full.sh
./scripts/test/smoke-discovery-e2e.sh --up && ./scripts/test/smoke-discovery-e2e.sh
```

### MCP stdio (Cursor / Claude)

| Server | Launcher | Example config |
|--------|----------|----------------|
| veil-mcp | [run-veil-mcp.sh](scripts/mcp/run-veil-mcp.sh) | [cursor.mcp.json.example](examples/mcp/cursor.mcp.json.example) |
| veil-engage | [run-veil-engage.sh](scripts/mcp/run-veil-engage.sh) | [engage.stdio.json.example](examples/mcp/engage.stdio.json.example) |

---

## Documentation

### Essential

| Document | Contents |
|----------|----------|
| [AGENTS.md](AGENTS.md) | Agent workflow, tests, core47 quick path |
| [docs/threatintel-runtime.md](docs/threatintel-runtime.md) | Compose, ports, NATS, bootstrap |
| [docs/mcp-agents.md](docs/mcp-agents.md) | veil-mcp + veil-engage setup |
| [deploy/README.md](deploy/README.md) | Layer compose, scaling, smokes |
| [docs/engage-tools.md](docs/engage-tools.md) | Catalog KPIs, matrices, assessment API |
| [docs/engage-lab-pentest.md](docs/engage-lab-pentest.md) | Install + self-pentest + HexStrike lab results |

### Engage depth

[engage/README.md](engage/README.md) · [engage-runtime.md](docs/engage-runtime.md) · [engage-install-linux.md](docs/engage-install-linux.md) · [engage-hardening.md](docs/engage-hardening.md) · [engage-audit-report.md](docs/engage-audit-report.md) · [engage-red-blue-lab.md](docs/engage-red-blue-lab.md)

### Reference

[platform-architecture.md](docs/platform-architecture.md) · [platform-closed-loop-pilot.md](docs/platform-closed-loop-pilot.md) · [deploy-platform-hybrid.md](docs/deploy-platform-hybrid.md) · [deploy-secure.md](docs/deploy-secure.md) · [ontology-appsec.md](docs/ontology-appsec.md) · [agent-evaluation-gaia.md](docs/agent-evaluation-gaia.md) · [external-security-frameworks.md](docs/external-security-frameworks.md)

**Layer READMEs:** [discovery/](discovery/README.md) · [pipeline/](pipeline/README.md) · [knowledge/](knowledge/README.md) · [scripts/](scripts/README.md)

---

## Tests

Full matrix: run from repo root. CI: [platform.yml](.github/workflows/platform.yml), [engage.yml](.github/workflows/engage.yml), [agent-eval.yml](.github/workflows/agent-eval.yml).

| Area | Commands |
|------|----------|
| **Shared / platform** | `make test-pkg-shared` · `make test-platform-p7` · `make test-platform-p0` · `make test-platform-unified-edge` · `make test-platform-closed-loop` · optional `make test-platform-full-loop` |
| **Layers** | `make test-discovery` · `make test-pipeline` · `make test-knowledge` · `make test-knowledge-serve` · `make test-graph-read-smoke` |
| **Engage gates** | `make test-engage` · `make test-engage-parity` · `make test-engage-executable-matrix` (**P9f**) · `make test-engage-hardening` · `make test-engage-events-pipeline` |
| **Eval / deploy** | `make test-agent-eval-pilot` · `make deploy-helm-template` · `make deploy-ansible-check` |

PR minimum: see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Graph packs

[docs/graph-pack.md](docs/graph-pack.md) · default **v0.4.6** in [versions.env](versions.env).

```bash
make graph-pack-export   # Neo4j must be running
make graph-pack-build
```

---

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n)[0] AS label, count(*) AS c ORDER BY c DESC LIMIT 20;
MATCH (v:Vulnerability)-[:HAS_CWE]->() RETURN count(*) AS has_cwe;
```
