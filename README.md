# Veil (Vulnerability Exploitation Intelligence Layer)

![Veil](docs/veil.png)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Veil** is a Neo4j-backed threat-intelligence platform with an optional **active security testing** layer. The graph stores vulnerabilities (CVE, CWE, CPE), LOLbins-style artifacts, detection content (Sigma/YARA/Caldera), TI feeds, SBOM advisories, and code-rule templates. Runtime is split into **four isolated Go contexts** — **discovery**, **pipeline**, **knowledge** (read intel), and **engage** (tool execution) — connected by **NATS JetStream** for ingestion and **dual MCP** servers for AI agents.

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Agents / AI:** [AGENTS.md](AGENTS.md) · **Security:** [SECURITY.md](SECURITY.md) · **Code of conduct:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## What you get

| Capability | Description |
|------------|-------------|
| **Threat graph** | Neo4j knowledge graph with versioned [graph packs](docs/graph-pack.md), HTTP API, and read-only MCP |
| **Ingestion bus** | Scrape → normalize (NED) → ingest over NATS with idempotent envelopes (`pkg/harvest`, `pkg/commit`) |
| **Engage toolkit** | **158/158** catalog · **54** bridge handlers · **104** subprocess-in-runner (`tools.live.yaml`, P10b python). Callable matrix: `make test-engage-executable-matrix`. Workflows (CTF, BB, CVE), Docker sandbox |
| **Closed loop** | Tool runs and findings → `engage.events` → graph (`EngageToolRun`, `EngageFinding`) for “act → learn → decide” |
| **Agent-ready** | Separate MCP: **veil-mcp** (read) vs **veil-engage** (exec), Keycloak RBAC, GAIA-style eval harness |
| **Prod path** | Hybrid deploy: Terraform + Ansible + Helm; secure overlays; control catalog (JCSF/DAF/OWASP-aligned) |

## Architecture

```mermaid
flowchart LR
  subgraph discovery [discovery]
    SW[scrape_worker]
    SW -->|harvest| NATS1[NATS SCRAPE]
  end
  subgraph pipeline [pipeline]
    PW[pipeline_worker]
    EEW[engage_events_worker]
    NATS1 --> PW
    PW -->|commit| NATS2[NATS INGEST]
    EEW -->|commit| NATS2
  end
  subgraph knowledge [knowledge read]
    IW[ingest_worker]
    Neo4j[(Neo4j)]
    API[veil-api]
    MCPg[veil-mcp]
    NATS2 --> IW --> Neo4j
    API --> Neo4j
    MCPg --> API
  end
  subgraph engage [engage exec]
    EngAPI[engage-api]
    EngMCP[veil-engage]
    Runner[engage-runner]
    EngMCP --> EngAPI
    EngAPI --> Runner
    EngAPI -.->|JWT read| API
    EngAPI -.->|optional| NATS_E[engage.events]
    NATS_E -.-> EEW
  end
```

| Layer | Path | Role | Agent MCP |
|-------|------|------|-----------|
| **Discovery** | [discovery/](discovery/) | Fetch feeds, Vitess ledger, publish `harvest` | — |
| **Pipeline** | [pipeline/](pipeline/) | NED → `commit`; [engage-events/](pipeline/engage-events/) bridges `engage.events.>` → `ingest.engage.*` | — |
| **Knowledge** | [knowledge/](knowledge/) | MERGE into Neo4j; [serve/](knowledge/serve/) read API + MCP | `veil-mcp` (read-only) |
| **Engage** | [engage/](engage/) | Catalog-driven tool execution, workflows, reports (hardened) | `veil-engage` (exec) |

**Shared contracts** (importable from any layer): [pkg/harvest](pkg/harvest/), [pkg/commit](pkg/commit/), [pkg/natsjet](pkg/natsjet/), [pkg/auth](pkg/auth/), [pkg/engage](pkg/engage/) (events, hostnorm, tool IDs), plus v8 shared logic: [pkg/report](pkg/report/), [pkg/decision](pkg/decision/), [pkg/exec](pkg/exec/), [pkg/api](pkg/api/), [pkg/mcp](pkg/mcp/). **No Go imports** across `discovery/`, `pipeline/`, `knowledge/`, `engage/`.

Deploy: [deploy/](deploy/) · Contracts: [docs/ingest-contract.md](docs/ingest-contract.md) · Graph: [docs/threatintel-runtime.md](docs/threatintel-runtime.md) · Engage: [docs/engage-runtime.md](docs/engage-runtime.md) · **Hybrid prod:** [docs/deploy-platform-hybrid.md](docs/deploy-platform-hybrid.md)

**Dual MCP for agents:** **`veil-graph`** (query TI) and **`veil-engage`** (run tools) — separate processes and RBAC. Legacy HexStrike (Python `:8888`) is **decommissioned** — see [docs/mcp-agents.md](docs/mcp-agents.md) and [docs/engage-audit-report.md](docs/engage-audit-report.md).

## Platform status

| Track | Status | Entry |
|-------|--------|--------|
| **Engage / HexStrike** | **P10 sign-off** — decommission `:8888`, **158/158** catalog, **54** bridge, **104** subprocess, route/NA gates green. Optional hardening: [engage_hexstrike_post_p10_signoff.plan.md](.cursor/plans/engage_hexstrike_post_p10_signoff.plan.md) | [engage-audit-report.md](docs/engage-audit-report.md) |
| **Platform v3–v4** | P0–P4b **done** — bus tests, closed/full loop, Terraform | [platform-full-loop-smoke.md](docs/platform-full-loop-smoke.md) |
| **Platform P5** | Hybrid deploy skeleton (TF + Ansible + Helm) | [deploy-platform-hybrid.md](docs/deploy-platform-hybrid.md) |
| **Platform P6** | **Done** — events, auth, scrapepub, stacks, natsjet publish | [veil_platform_refactor_p6.plan.md](.cursor/plans/veil_platform_refactor_p6.plan.md) |
| **Platform P7** | **Done** — `pkg/*/domain`, `test-platform-p7` CI | [domain-contour.md](docs/domain-contour.md) |
| **Platform v8** | **Done** — layer renames, `pkg/report`, `pkg/decision`, `pkg/exec`, `pkg/api`, `pkg/mcp`, browser → discovery | [platform-architecture.md](docs/platform-architecture.md), [v8 master plan](.cursor/plans/veil_platform_v8_layers_master.plan.md) |
| **Platform P12** | **In progress** — single TLS edge, path routing, stateless scale **4/8/16**, Neo4j Enterprise 3-core (prod) | [platform-unified-access.md](docs/platform-unified-access.md), [P12 master plan](.cursor/plans/veil_platform_p12_unified_access.plan.md) |
| **Security** | veil-controls + engage hardening; prod pentest 0 HIGH | [external-security-frameworks.md](docs/external-security-frameworks.md) |
| **Agent eval** | GAIA offline harness | [agent-evaluation-gaia.md](docs/agent-evaluation-gaia.md) |

## Platform v8 (done)

| Logical layer | Path | Shared `pkg/` |
|---------------|------|----------------|
| **Discovery** | [discovery/](discovery/) | `pkg/exec` (optional fetcher spike), [discovery/pkg/browser](discovery/pkg/browser/) |
| **Pipeline** | [pipeline/](pipeline/) | `pkg/ti/*`, `pkg/commit` |
| **Knowledge** | [knowledge/](knowledge/) | `pkg/api`, `pkg/mcp` (transport); read via veil-api |
| **Engage** | [engage/](engage/) | `pkg/report`, `pkg/decision`, `pkg/exec` (runner adapter) |
| **Report** | — | [pkg/report](pkg/report/) |
| **API + MCP** | veil-api / veil-mcp / veil-engage | [pkg/api](pkg/api/), [pkg/mcp](pkg/mcp/) |

NATS wire (`scrape.>`, `ingest.>`) and product names (veil-api, veil-mcp) unchanged. Master plan (all phases merged): [veil_platform_v8_layers_master.plan.md](.cursor/plans/veil_platform_v8_layers_master.plan.md).

## Quick start

### Graph only (demo API + optional pack import)

```bash
docker compose up --build -d
```

| Endpoint | Default (dev compose) |
|----------|------------------------|
| Neo4j Browser | http://localhost:7474 (`neo4j` / `neo4jpassword`) |
| HTTP API | http://localhost:8090 |
| MCP Streamable HTTP | http://localhost:8091/mcp (`--profile mcp`, `MCP_HTTP_ENABLED=1`) |

Production secure overlay (TLS on **443** only, no published Neo4j): [docs/deploy-secure.md](docs/deploy-secure.md).

**Unified access (P12):** one TLS hostname routes graph and engage by path — `/v1/*`, `/api/*`, `/mcp/graph`, `/mcp/engage`. Dev may still use direct ports (`8090`, `8091`, `8890`, `8892`). Stateless tiers scale with `VEIL_API_SCALE`, `VEIL_MCP_GRAPH_SCALE`, `VEIL_ENGAGE_API_SCALE`, `VEIL_MCP_ENGAGE_SCALE` (`4` | `8` | `16`). Prod Neo4j uses Enterprise **3-core** cluster; local/CI stays single `neo4j:5`. Operator contract: [docs/platform-unified-access.md](docs/platform-unified-access.md).

`graph-bootstrap` imports the default graph pack ([versions.env](versions.env) → `GRAPH_PACK_VERSION`, currently **v0.4.6**) when published, unless `GRAPH_PACK_SKIP=1`.

```bash
curl -sS http://localhost:8090/health
curl -sS http://localhost:8090/v1/categories | jq .
```

### Engage layer (opt-in offensive tooling)

```bash
docker compose -f deploy/engage/compose.yml up -d --build engage-api engage-mcp
curl -sS http://localhost:8890/health | jq .
curl -sS http://localhost:8890/api/tools | jq .
```

| Service | Port | Notes |
|---------|------|--------|
| engage-api | 8890 | `POST /api/tools/{name}`, intelligence, workflows |
| veil-engage MCP | stdio or :8892 | [engage.stdio.json.example](examples/mcp/engage.stdio.json.example) |
| engage-runner | — | `docker compose --profile runner` + `ENGAGE_RUNNER_MODE=docker` |

Docs: [engage/README.md](engage/README.md) · [engage-hardening.md](docs/engage-hardening.md) · [engage-legacy-parity.md](docs/engage-legacy-parity.md). Catalog tools prefixed `ai_*` are **not** backed by an LLM today — see [engage-llm-stubs.md](docs/engage-llm-stubs.md) for stub vs intel-bridge dispatch and planned future provider scope.

**Events bus (optional):** tool runs and findings publish to NATS when `ENGAGE_EVENTS_NATS_ENABLED=1`; pipeline bridges to `ingest.engage.*` and graph ingest persists `EngageToolRun` / `EngageFinding` nodes.

```bash
docker compose -f deploy/engage/compose.yml -f deploy/engage/compose.events.yml \
  up -d --build nats engage-api engage-events-worker
make test-engage-events-pipeline
```

### Full scrape pipeline

```bash
./scripts/ops/compose-up-full.sh
```

E2E smoke: `./scripts/test/smoke-scrape-e2e.sh --up` then `./scripts/test/smoke-scrape-e2e.sh`

### Platform integration smokes

```bash
make test-platform-p0              # pkg + NATS bus (no Docker)
make test-platform-closed-loop     # engage → ingest → Neo4j → target-graph
make test-platform-full-loop         # scrape + closed loop (heavy Docker)
```

CI: [platform.yml](.github/workflows/platform.yml), [engage.yml](.github/workflows/engage.yml), [agent-eval.yml](.github/workflows/agent-eval.yml).

### Production deploy (hybrid)

| Layer | Path | Role |
|-------|------|------|
| Terraform | [deploy/terraform/](deploy/terraform/README.md) | Cloud foundation + compose env |
| Ansible | [deploy/ansible/](deploy/ansible/README.md) | Data plane VMs (Neo4j, NATS, scrape cron) |
| Helm | [deploy/helm/veil/](deploy/helm/veil/README.md) | Control plane on K8s (api, engage, MCP HTTP, workers, HPA; scale **4/8/16**) |
| Profiles | [deploy/profiles/](deploy/profiles/) | `secure-engage.env`, stack env overlays |
| Controls | [deploy/security/veil-controls.yaml](deploy/security/veil-controls.yaml) | Machine-readable security catalog |

```bash
make deploy-helm-template    # render chart (needs helm)
make deploy-ansible-check    # syntax-check playbooks (needs ansible)
make sync-github-metadata    # push .github/repo-description.txt → GitHub
```

## AI agents & automation

| Artifact | Purpose |
|----------|---------|
| [AGENTS.md](AGENTS.md) | Rules for Cursor/CI bots (layers, critic, docs) |
| [.cursor/agents/manifest.yaml](.cursor/agents/manifest.yaml) | Declarative subagent definitions (`make agents-render`) |
| [docs/mcp-agents.md](docs/mcp-agents.md) | veil-mcp + veil-engage setup |
| [eval/gaia/](eval/gaia/) | GAIA-aligned eval harness ([docs/agent-evaluation-gaia.md](docs/agent-evaluation-gaia.md)) |
| [docs/external-agent-store.md](docs/external-agent-store.md) | openJiuwen Agent Store (reference; `make external-clone-agent-store`) |

## Documentation index

| Document | Contents |
|----------|----------|
| [AGENTS.md](AGENTS.md) | Cursor/agents: read [coding-style.md](docs/coding-style.md) first |
| [docs/threatintel-runtime.md](docs/threatintel-runtime.md) | Compose, ports, env, bootstrap, graph API/MCP, NATS |
| [docs/engage-runtime.md](docs/engage-runtime.md) | Engage API/MCP, runner isolation, RBAC |
| [docs/deploy-secure.md](docs/deploy-secure.md) | Prod hardening: nginx TLS, distroless, auth fail-closed |
| [docs/auth-keycloak.md](docs/auth-keycloak.md) | Optional JWT + RBAC for API and MCP |
| [deploy/README.md](deploy/README.md) | Per-layer compose, scaling, smoke, graph pack releases |
| [docs/deploy-platform-hybrid.md](docs/deploy-platform-hybrid.md) | P5: Terraform + Ansible + Helm |
| [docs/platform-closed-loop-pilot.md](docs/platform-closed-loop-pilot.md) | Act → learn → remember → decide |
| [docs/platform-full-loop-smoke.md](docs/platform-full-loop-smoke.md) | Scrape + closed loop (P4b) |
| [docs/engage-audit-report.md](docs/engage-audit-report.md) | HexStrike migration sign-off |
| [docs/engage-hardening.md](docs/engage-hardening.md) | Active-defense hardening + safe self-test |
| [docs/engage-agentic-threats.md](docs/engage-agentic-threats.md) | Agentic AI / MCP threats ↔ mitigations |
| [docs/external-security-frameworks.md](docs/external-security-frameworks.md) | JCSF / DAF / OWASP → Veil controls |
| [docs/external-agent-store.md](docs/external-agent-store.md) | Agent Store reference patterns |
| [docs/agent-evaluation-gaia.md](docs/agent-evaluation-gaia.md) | GAIA eval (arXiv primary; HF optional) |
| [discovery/README.md](discovery/README.md) | Discovery sources and env vars |
| [pipeline/README.md](pipeline/README.md) | Pipeline worker and normalization |
| [knowledge/README.md](knowledge/README.md) | Ingest, API, MCP, Neo4j client |
| [engage/README.md](engage/README.md) | Tool catalog, veil-engage MCP, workflows |
| [docs/platform-architecture.md](docs/platform-architecture.md) | Current + v8 layers, runner vs factory |
| [docs/platform-unified-access.md](docs/platform-unified-access.md) | P12: single TLS edge, path map, scale 4/8/16, Neo4j cluster |
| [docs/domain-contour.md](docs/domain-contour.md) | pkg domain SOT map |
| [docs/coding-style.md](docs/coding-style.md) | Architecture, four contexts, PR checklist |
| [docs/mcp-agents.md](docs/mcp-agents.md) | veil-graph + veil-engage agent setup |
| [docs/engage-tools.md](docs/engage-tools.md) | Catalog YAML, parameters, enable-by-category |
| [scripts/README.md](scripts/README.md) | Export, packs, smoke, engage scripts |

## MCP (agents)

| MCP server | Layer | Transport | Example |
|------------|-------|-----------|---------|
| **veil-mcp** | Graph read | stdio / HTTP :8091 or edge `/mcp/graph` | [run-veil-mcp.sh](scripts/mcp/run-veil-mcp.sh) |
| **veil-engage** | Tool exec | stdio / HTTP :8892 or edge `/mcp/engage` | [run-veil-engage.sh](scripts/mcp/run-veil-engage.sh) |

**Remote / prod:** one TLS host — graph MCP at `https://<veil-host>/mcp/graph`, engage MCP at `https://<veil-host>/mcp/engage` (stdio stays two processes). See [docs/platform-unified-access.md](docs/platform-unified-access.md).

Setup: [docs/mcp-agents.md](docs/mcp-agents.md). Keycloak: [docs/auth-keycloak.md](docs/auth-keycloak.md). Examples: [examples/mcp/](examples/mcp/).

## Graph packs

See [docs/graph-pack.md](docs/graph-pack.md).

## Tests

```bash
make test-pkg-shared              # pkg/harvest, commit, natsjet, auth, engage/events
make test-discovery
make test-pipeline
make test-knowledge                   # graph modules + serve build
make test-knowledge-serve             # knowledge/serve unit tests (-race)
make test-graph-read-smoke        # Docker: Neo4j + API + MCP HTTP
make test-engage                  # engage layer unit tests + build
make test-engage-parity           # catalog 150 tools vs legacy MCP reference
make test-engage-hardening        # safe self-test + veil-controls audit
make test-engage-secure           # Docker TLS overlay smoke
make test-engage-compose          # Docker: async jobs + runner profile
make test-engage-events-pipeline  # Docker: engage.events → ingest.engage.*
make test-agent-eval-pilot        # GAIA offline harness smoke
make test-agent-eval-paper        # GAIA arXiv Fig. 1 format checks
make test-platform-p7             # pkg domain + bus slices (CI on PR)
make test-platform-p0             # Platform bus unit tests
make test-platform-closed-loop    # Platform closed-loop pilot (Docker)
make test-platform-full-loop      # Platform full loop with scrape (Docker, heavy)
make deploy-helm-template         # Helm chart render check
make sync-github-metadata         # Update GitHub description from .github/repo-description.txt
make pentest-veil-dual            # Docker target + safe self-pentest report
make pentest-veil-mcp             # Pentest only (stack must be up); report in eval/results/
```

## Smoke Cypher

```cypher
MATCH (n) RETURN labels(n)[0] AS label, count(*) AS c ORDER BY c DESC LIMIT 20;
MATCH (v:Vulnerability)-[:HAS_CWE]->() RETURN count(*) AS has_cwe;
```
