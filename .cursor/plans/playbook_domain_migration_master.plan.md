---
name: Playbook domain migration master
overview: "Structured import of Anthropic skills into pkg/playbook — procedure index, ontology API, graph props, engage catalog bridge. Tracks subdomain matrix in docs/playbooks/playbook-import-matrix.md."
todos:
  - id: i0-matrix-schema
    content: "I0: docs/playbooks/playbook-import-matrix.md + ProcedureSpec + extract-procedures-index.py + CI"
    status: completed
  - id: o1-ontology
    content: "O1-O3: pkg/playbook/framework subdomain registry + ontology API"
    status: completed
  - id: k1-procedure-api
    content: "K1-K3: pkg/playbook/procedure + veil-api/MCP structured read"
    status: completed
  - id: g1-graph-props
    content: "G1-G3: CyberSkill graph props + playbook_seed stepCount"
    status: completed
  - id: e1-engage-bridge
    content: "E1-E3: cataloglink boost + engage procedure hints"
    status: completed
  - id: migration-master
    content: "Master plan status table"
    status: completed
isProject: true
---

# Playbook domain migration — master plan

Prerequisite: [cyber_corpus_vendor_master.plan.md](cyber_corpus_vendor_master.plan.md) (corpus committed).

## Status

| Phase | Branch | Status |
|-------|--------|--------|
| I0 schema + matrix | `feat/playbook-i0-procedure-schema` | **done** |
| O1–O3 ontology | `feat/pkg-playbook-ontology` | **done** |
| K1–K3 procedure API | `feat/knowledge-playbook-procedure-api` | **done** |
| G1–G3 graph props | `feat/knowledge-playbook-graph-props` | **done** (run `playbook_seed` locally for edges) |
| E1–E3 engage bridge | `engage/phase-playbook-procedure-bridge` | **done** |
| V5 native YAML | per [playbook-import-matrix.md](../../docs/playbooks/playbook-import-matrix.md) | **pending** (long tail) |

## Verify

```bash
make procedures-index check-procedures-index
cd pkg && go test ./playbook/...
make test-knowledge-serve
cd engage/serve && env GOWORK=$(dirname $(pwd))/go.work go test ./internal/usecase/intelligence/... -count=1 -run 'Playbook|Procedure'
```

Detail: [playbook_domain_migration_e7c6318a.plan.md](playbook_domain_migration_e7c6318a.plan.md) (do not edit).
