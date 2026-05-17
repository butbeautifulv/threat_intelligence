---
name: Veil P12 unified access
overview: "Единый edge (nginx + gateway): veil-api/veil-mcp, scale 4/8/16, Neo4j Enterprise cluster."
todos:
  - id: p12a-adr
    content: "P12a: ADR docs/platform-unified-access.md"
    status: in_progress
  - id: p12b-nginx
    content: "P12b: deploy/platform/nginx unified edge"
    status: in_progress
  - id: p12c-scale
    content: "P12c: VEIL_*_SCALE + compose-scale-veil.sh"
    status: in_progress
  - id: p12d-neo4j
    content: "P12d: Neo4j Enterprise 3-core compose"
    status: in_progress
  - id: p12e-gateway
    content: "P12e: platform/gateway skeleton"
    status: in_progress
  - id: p12f-mcp
    content: "P12f: unified MCP HTTP aggregator"
    status: in_progress
  - id: p12g-api-facade
    content: "P12g: composite health + proxy façade"
    status: in_progress
  - id: p12h-auth
    content: "P12h: secure-unified stack"
    status: in_progress
  - id: p12i-smoke
    content: "P12i: smoke-unified-edge + CI"
    status: in_progress
  - id: p12j-docs
    content: "P12j: README, mcp-agents, AGENTS"
    status: completed
isProject: false
---

# P12 — Unified access (master)

See full design: [veil_unified_access_p12_438f3804.plan.md](veil_unified_access_p12_438f3804.plan.md) (Cursor plan artifact).

Operator contract (ADR): [docs/platform-unified-access.md](../../docs/platform-unified-access.md).

## Merge order

`P12a` → `P12d` → `P12b` → `P12c` → `P12e` → `P12g` → `P12f` → `P12h` → `P12i` → `P12j`

## Branches (status)

| Phase | Branch | Status | Owner |
|-------|--------|--------|-------|
| P12a | `platform/p12a-unified-access-adr` | in_progress | — |
| P12b | `platform/p12b-veil-nginx-edge` | in_progress | — |
| P12c | `platform/p12c-stateless-scale` | in_progress | — |
| P12d | `platform/p12d-neo4j-enterprise-cluster` | in_progress | — |
| P12e | `platform/p12e-gateway-skeleton` | in_progress | — |
| P12f | `platform/p12f-unified-mcp-http` | in_progress | — |
| P12g | `platform/p12g-unified-api-facade` | in_progress | — |
| P12h | `platform/p12h-unified-auth-edge` | in_progress | — |
| P12i | `platform/p12i-unified-edge-smoke` | in_progress | — |
| P12j | `platform/p12j-unified-access-docs` | **done** | operator docs merged on branch |

## Verification

```bash
make test-platform-unified-edge
make test-platform-p7
make test-knowledge test-engage
```
