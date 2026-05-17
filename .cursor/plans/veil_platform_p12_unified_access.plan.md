---
name: Veil P12 unified access
overview: "Единый edge (nginx + gateway): veil-api/veil-mcp, scale 4/8/16, Neo4j Enterprise cluster."
todos:
  - id: p12a-adr
    content: "P12a: ADR docs/platform-unified-access.md"
    status: completed
  - id: p12b-nginx
    content: "P12b: deploy/platform/nginx unified edge"
    status: pending
  - id: p12c-scale
    content: "P12c: VEIL_*_SCALE + compose-scale-veil.sh"
    status: pending
  - id: p12d-neo4j
    content: "P12d: Neo4j Enterprise 3-core compose"
    status: pending
  - id: p12e-gateway
    content: "P12e: platform/gateway skeleton"
    status: pending
  - id: p12f-mcp
    content: "P12f: unified MCP HTTP aggregator"
    status: pending
  - id: p12g-api-facade
    content: "P12g: composite health + proxy façade"
    status: pending
  - id: p12h-auth
    content: "P12h: secure-unified stack"
    status: pending
  - id: p12i-smoke
    content: "P12i: smoke-unified-edge + CI"
    status: pending
  - id: p12j-docs
    content: "P12j: README, mcp-agents, AGENTS"
    status: pending
isProject: false
---

# P12 — Unified access (master)

**ADR:** [docs/platform-unified-access.md](../../docs/platform-unified-access.md)  
**Architecture:** [docs/platform-architecture.md](../../docs/platform-architecture.md)

## Merge order

`P12a` → `P12d` → `P12b` → `P12c` → `P12e` → `P12g` → `P12f` → `P12h` → `P12i` → `P12j`

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P12a | `platform/p12a-unified-access-adr` | done — ADR + arch mermaid |
| P12b | `platform/p12b-veil-nginx-edge` | pending |
| P12c | `platform/p12c-stateless-scale` | pending |
| P12d | `platform/p12d-neo4j-enterprise-cluster` | pending |
| P12e | `platform/p12e-gateway-skeleton` | pending |
| P12f | `platform/p12f-unified-mcp-http` | pending |
| P12g | `platform/p12g-unified-api-facade` | pending |
| P12h | `platform/p12h-unified-auth-edge` | pending |
| P12i | `platform/p12i-unified-edge-smoke` | pending |
| P12j | `platform/p12j-unified-access-docs` | pending |

## Verification

```bash
make test-platform-unified-edge
make test-platform-p7
make test-knowledge test-engage
```
