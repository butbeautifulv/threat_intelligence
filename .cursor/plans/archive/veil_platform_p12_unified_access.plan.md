---
name: Veil P12 unified access
overview: "Единый edge (nginx + gateway): veil-api/veil-mcp, scale 4/8/16, Neo4j Enterprise cluster."
todos:
  - id: p12a-adr
    content: "P12a: ADR docs/architecture/platform-unified-access.md"
    status: completed
  - id: p12b-nginx
    content: "P12b: deploy/platform/nginx unified edge"
    status: completed
  - id: p12c-scale
    content: "P12c: VEIL_*_SCALE + compose-scale-veil.sh"
    status: completed
  - id: p12d-neo4j
    content: "P12d: Neo4j Enterprise 3-core compose"
    status: completed
  - id: p12e-gateway
    content: "P12e: platform/gateway skeleton"
    status: completed
  - id: p12f-mcp
    content: "P12f: unified MCP HTTP aggregator"
    status: completed
  - id: p12g-api-facade
    content: "P12g: composite health + proxy façade"
    status: completed
  - id: p12h-auth
    content: "P12h: secure-unified stack"
    status: completed
  - id: p12i-smoke
    content: "P12i: smoke-unified-edge + CI"
    status: completed
  - id: p12j-docs
    content: "P12j: README, mcp-agents, AGENTS"
    status: completed
isProject: false
---

# P12 — Unified access (master)

**ADR:** [docs/architecture/platform-unified-access.md](../../docs/architecture/platform-unified-access.md)  
**Architecture:** [docs/architecture/platform-architecture.md](../../docs/architecture/platform-architecture.md)

## Merge order

`P12a` → `P12d` → `P12b` → `P12c` → `P12e` → `P12g` → `P12f` → `P12h` → `P12i` → `P12j`

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| P12a | `platform/p12a-unified-access-adr` | done |
| P12b | `platform/p12b-veil-nginx-edge` | done |
| P12c | `platform/p12c-stateless-scale` | done |
| P12d | `platform/p12d-neo4j-enterprise-cluster` | done |
| P12e | `platform/p12e-gateway-skeleton` | done |
| P12f | `platform/p12f-unified-mcp-http` | done |
| P12g | `platform/p12g-unified-api-facade` | done |
| P12h | `platform/p12h-unified-auth-edge` | done |
| P12i | `platform/p12i-unified-edge-smoke` | done — `scripts/test/smoke-unified-edge.sh`, CI `platform.yml` |
| P12j | `platform/p12j-unified-access-docs` | done |

## Verification

```bash
make test-platform-unified-edge
make test-platform-p7
make test-knowledge test-engage
```

## Sign-off (2026-05)

All phases P12a–P12j merged to `main`. Release gate for operators:

- `make test-platform-unified-edge` — Docker smoke through `veil-edge` (CI job `platform.yml` → `unified-edge`)
- ADR and path map: [docs/architecture/platform-unified-access.md](../../docs/architecture/platform-unified-access.md)
- Dev TLS: [deploy/platform/nginx/certs/README.md](../../deploy/platform/nginx/certs/README.md) (do not commit keys)
